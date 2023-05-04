package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"runtime"
	"sync"

	"github.com/pkg/errors"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/rancherlabs/corral/pkg/shell"
	"github.com/rancherlabs/corral/pkg/vars"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

var (
	cfgViper = viper.New()
)

const createDescription = `
Create a new corral from the given package. Packages can either be a valid OCI reference or a path to a local directory.

Examples:
corral create k3s ghcr.io/rancher/k3s
corral create k3s-ha -v controlplane_count=3 ghcr.io/rancher/k3s
corral create k3s-custom /home/rancher/issue-1234
`

func NewCommandCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME PACKAGE",
		Short: "Create a new corral",
		Long:  createDescription,
		Args:  cobra.RangeArgs(1, 2),
		RunE:  create,
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfgFile := cmd.Flags().Lookup("config").Value.String()
			if cfgFile != "" {
				cfgViper.AddConfigPath(cfgFile)
				err := cfgViper.ReadInConfig()
				if err != nil {
					logrus.Fatalf("Error reading config file: %v", err)
				}
			}
		},
	}

	cmd.Flags().String("config", "", "loadManifest flags for this command from a file.")

	cmd.Flags().StringArrayP("variable", "v", []string{}, "Set a variable to configure the package.")
	_ = cfgViper.BindPFlag("variable", cmd.Flags().Lookup("variable"))

	cmd.Flags().StringP("package", "p", "", "Set a variable to configure the package.")
	_ = cfgViper.BindPFlag("package", cmd.Flags().Lookup("package"))

	cmd.Flags().Bool("recreate", false, "Destroy corral with the same name if it exists before creating.")
	_ = cfgViper.BindPFlag("recreate", cmd.Flags().Lookup("recreate"))

	cmd.Flags().Bool("skip-cleanup", false, "Do not run terraform destroy when an error is encountered. This can result in un-tracked infrastructure resources!")
	_ = cfgViper.BindPFlag("skip-cleanup", cmd.Flags().Lookup("skip-cleanup"))

	return cmd
}

func create(cmd *cobra.Command, args []string) error {
	cfg := config.MustLoad()

	var corr corral.Corral
	corr.RootPath = config.CorralPath(args[0])
	corr.Name = args[0]
	corr.Source = cfgViper.GetString("package")
	corr.NodePools = map[string][]corral.Node{}
	corr.Vars = map[string]any{}

	if len(args) > 1 {
		corr.Source = args[1]
	}
	// get the source from flags or args
	if corr.Source == "" {
		logrus.Fatal("You must specify a package with the `-p` flag or as an argument.")
	}

	if cfgViper.GetBool("recreate") {
		logrus.Infof("Deleting existing corral [%s]", args[0])
		deleteCorrals(cmd, args[0:1])
	}

	// ensure this corral is unique
	if corr.Exists() {
		logrus.Fatalf("corral [%s] already exists", corr.Name)
	}

	// load cli variables
	for _, raw := range cfgViper.GetStringSlice("variable") {
		k, v, err := vars.ToVar(raw)
		if err != nil {
			return err
		}
		if k == "" {
			logrus.Fatal("variables should be in the format <key>=<value>")
		}
		corr.Vars[k] = v
	}
	for k, v := range cfg.Vars { // copy the global vars for future reference
		corr.Vars[k] = v
	}

	// load the package
	logrus.Info("loading package")
	pkg, err := _package.LoadPackage(corr.Source)
	if err != nil {
		logrus.Fatalf("failed to load package: %s", err)
	}

	// update the corral ref to the absolute path
	corr.Source = pkg.RootPath

	// validate the variables
	err = pkg.ValidateVarSet(corr.Vars, true)
	if err != nil {
		logrus.Fatal("invalid variables: ", err)
	}

	err = pkg.ApplyDefaultVars(corr.Vars)
	if err != nil {
		logrus.Fatal("invalid defaults: ", err)
	}

	if corr.Vars["corral_private_key"] == nil && corr.Vars["corral_public_key"] == nil {
		logrus.Info("generating ssh keys")
		privkey, _ := generatePrivateKey(2048)
		pubkey, _ := generatePublicKey(&privkey.PublicKey)
		corr.PrivateKey = string(encodePrivateKeyToPEM(privkey))
		corr.PublicKey = string(pubkey)
		corr.Vars["corral_public_key"] = corr.PublicKey
		corr.Vars["corral_private_key"] = corr.PrivateKey
	} else {
		logrus.Info("reusing generated ssh keys")
		corr.PublicKey = corr.Vars["corral_public_key"].(string)
		corr.PrivateKey = corr.Vars["corral_private_key"].(string)
	}
	// add common variables
	userPublicKey, err := os.ReadFile(cfg.UserPublicKeyPath)
	if err != nil {
		logrus.Error("failed to read user public key: ", err)
	}
	corr.Vars["corral_name"] = corr.Name
	corr.Vars["corral_user_id"] = cfg.UserID
	corr.Vars["corral_user_public_key"] = string(userPublicKey)
	corr.Vars["corral_node_pools"] = ""

	// write the corral to disk
	corr.SetStatus(corral.StatusProvisioning)

	var lastCommand int
	knownNodes := map[*shell.Shell]struct{}{}
	shellRegistry := shell.NewRegistry()
	for i, cmd := range pkg.Manifest.Commands {
		lastCommand = i

		if cmd.Module != "" {
			logrus.Infof("[%d/%d] applying %s module", i+1, len(pkg.Manifest.Commands), cmd.Module)
			err = corr.ApplyModule(pkg.TerraformModulePath(cmd.Module), cmd.Module)
			if err != nil {
				corr.SetStatus(corral.StatusError)
				break
			}
		}

		if cmd.Parallel == nil {
			cmd.Parallel = &[]bool{true}[0]
		}

		if cmd.Command != "" {
			logrus.Infof("[%d/%d] running command %s", i+1, len(pkg.Manifest.Commands), cmd.Command)

			// find all distinct nodes in the given node pools
			var shells []*shell.Shell
			seen := map[*shell.Shell]struct{}{}
			for _, name := range cmd.NodePoolNames {
				if np := corr.NodePools[name]; np != nil {
					for _, n := range np {
						// get or create a shell pointer for the node
						sh, err := shellRegistry.GetShell(n, corr.PrivateKey, corr.Vars)
						if err != nil {
							corr.SetStatus(corral.StatusError)
							logrus.Errorf("failed to connect to node [%s]: %s", n.Name, err)
							break
						}

						// add distinct shells to the shells list
						if _, ok := seen[sh]; !ok {
							seen[sh] = struct{}{}
							shells = append(shells, sh)
						}
					}
				}
			}

			err = executeShellCommand(cmd.Command, shells, corr.Vars, *cmd.Parallel)
		}

		if err != nil {
			corr.SetStatus(corral.StatusError)
			logrus.Error(err)
			break
		}

		// collect new nodes to copy files
		var newNodeShells []*shell.Shell
		for npName, np := range corr.NodePools {
			for _, n := range np {
				n.OverlayRoot = pkg.Overlay[npName]
				sh, err := shellRegistry.GetShell(n, corr.PrivateKey, corr.Vars)
				if err != nil {
					corr.SetStatus(corral.StatusError)
					logrus.Errorf("failed to connect to node [%s]: %s", n.Name, err)
					break
				}

				if _, ok := knownNodes[sh]; !ok {
					newNodeShells = append(newNodeShells, sh)
					knownNodes[sh] = struct{}{}
				}
			}
		}

		// copy package files to new nodes
		err = copyPackageFiles(newNodeShells, pkg)
		if err != nil {
			corr.SetStatus(corral.StatusError)
			logrus.Error("failed to copy package files: ", err)
			break
		}

		_ = corr.Save()
	}

	// close all shells
	shellRegistry.Close()

	// if the corral is in an error state delete it
	if corr.Status == corral.StatusError {
		if cfgViper.GetBool("skip-cleanup") {
			logrus.Warnf("skipping roll back")
			_ = corr.Save()
		} else {
			logrus.Info("attempting to roll back corral")
			for i := lastCommand; i >= 0; i-- {
				if pkg.Commands[i].Module != "" {
					if pkg.Commands[i].SkipCleanup {
						continue
					}

					logrus.Infof("rolling back %s module", pkg.Commands[i].Module)
					if err = corr.DestroyModule(pkg.Commands[i].Module); err != nil {
						logrus.Fatalf("failed to cleanup module [%s]: %v", pkg.Commands[i].Module, err)
					}
				}
			}
			_ = corr.Delete()
		}
	} else {
		corr.SetStatus(corral.StatusReady)
	}

	logrus.Info("done!")
	return nil
}

// copyPackageFiles copies the appropriate overlay files from the given package to the shells.  Concurrency is limited
// to the number of cpus on the user's machine.
func copyPackageFiles(shells []*shell.Shell, pkg _package.Package) error {
	var wg errgroup.Group
	sem := make(chan bool, runtime.NumCPU())

	for _, sh := range shells {
		sh := sh
		wg.Go(func() error {
			sem <- true

			err := sh.UploadPackageFiles(pkg)

			<-sem
			return err
		})
	}

	return wg.Wait()
}

func executeShellCommand(command string, shells []*shell.Shell, vs vars.VarSet, parallel bool) error {
	var err error
	if parallel {
		err = executeShellCommandAsync(command, shells, vs)
	} else {
		err = executeShellCommandSync(command, shells, vs)
	}
	if err != nil {
		return errors.Wrapf(err, "running %s", command)
	}
	return nil
}

// executeShellCommandAsync runs the given command on the given shells. Any vars set are saved to the VarSet.
// Concurrency is limited to the number of cpus on the user's machine.
func executeShellCommandAsync(command string, shells []*shell.Shell, vs vars.VarSet) error {
	var mu sync.Mutex
	var wg errgroup.Group
	sem := make(chan bool, runtime.NumCPU())

	for _, sh := range shells {
		sh := sh
		wg.Go(func() error {
			sem <- true

			err := sh.Run(command)
			if err != nil {
				<-sem
				return err
			}

			mu.Lock()
			for k, v := range sh.Vars {
				vs[k] = v
			}
			mu.Unlock()

			<-sem
			return nil
		})
	}

	return wg.Wait()
}

func executeShellCommandSync(command string, shells []*shell.Shell, vs vars.VarSet) error {
	for _, sh := range shells {
		sh := sh
		err := sh.Run(command)
		if err != nil {
			return err
		}

		for k, v := range sh.Vars {
			vs[k] = v
		}
	}

	return nil
}

func generatePrivateKey(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func generatePublicKey(key *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(key)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := bytes.Replace(ssh.MarshalAuthorizedKey(publicRsaKey), []byte("\n"), []byte(""), 2)

	return pubKeyBytes, nil
}

func encodePrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(key)

	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return pem.EncodeToMemory(&privBlock)
}
