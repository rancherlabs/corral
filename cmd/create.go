package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/rancherlabs/corral/pkg/vars"
	"os"
	"sync"
	"time"

	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/rancherlabs/corral/pkg/shell"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/wait"
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
		Run:   create,
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfgFile := cmd.Flags().Lookup("config").Value.String()
			if cfgFile != "" {
				cfgViper.AddConfigPath(cfgFile)
				err := cfgViper.ReadInConfig()
				if err != nil {
					logrus.Fatal("failed to parse config file.")
				}
			}
		},
	}

	cmd.Flags().String("config", "", "loadManifest flags for this command from a file.")

	cmd.Flags().StringArrayP("variable", "v", []string{}, "Set a variable to configure the package.")
	_ = cfgViper.BindPFlag("variable", cmd.Flags().Lookup("variable"))

	cmd.Flags().StringP("package", "p", "", "Set a variable to configure the package.")
	_ = cfgViper.BindPFlag("package", cmd.Flags().Lookup("package"))

	return cmd
}

func create(_ *cobra.Command, args []string) {
	cfg := config.MustLoad()

	var corr corral.Corral
	corr.RootPath = config.CorralPath(args[0])
	corr.Name = args[0]
	corr.Source = cfgViper.GetString("package")
	corr.NodePools = map[string][]corral.Node{}
	corr.Vars = map[string]string{}

	if len(args) > 1 {
		corr.Source = args[1]
	}
	// get the source from flags or args
	if corr.Source == "" {
		logrus.Fatal("You must specify a package with the `-p` flag or as an argument.")
	}

	// ensure this corral is unique
	if corr.Exists() {
		logrus.Fatalf("corral [%s] already exists", corr.Name)
	}

	// load cli variables
	for _, raw := range cfgViper.GetStringSlice("variable") {
		k, v := vars.ToVar(raw)
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

	logrus.Info("generating ssh keys")
	privkey, _ := generatePrivateKey(2048)
	pubkey, _ := generatePublicKey(&privkey.PublicKey)
	corr.PrivateKey = string(encodePrivateKeyToPEM(privkey))
	corr.PublicKey = string(pubkey)

	// add common variables
	userPublicKey, err := os.ReadFile(cfg.UserPublicKeyPath)
	if err != nil {
		logrus.Error("failed to read user public key: ", err)
	}
	corr.Vars["corral_name"] = corr.Name
	corr.Vars["corral_user_id"] = cfg.UserID
	corr.Vars["corral_user_public_key"] = string(userPublicKey)
	corr.Vars["corral_public_key"] = corr.PublicKey
	corr.Vars["corral_private_key"] = corr.PrivateKey
	corr.Vars["corral_node_pools"] = ""

	// write the corral to disk
	corr.SetStatus(corral.StatusProvisioning)

	var lastCommand int
	knownAddresses := map[string]struct{}{}
	for i, cmd := range pkg.Manifest.Commands {
		lastCommand = i

		if cmd.Module != "" {
			logrus.Infof("[%d/%d] applying %s module", i+1, len(pkg.Manifest.Commands), cmd.Module)
			err = corr.ApplyModule(pkg.TerraformModulePath(cmd.Module), cmd.Module)
		}

		if cmd.Command != "" {
			logrus.Infof("[%d/%d] running command %s", i+1, len(pkg.Manifest.Commands), cmd.Command)
			err = executeShellCommand(cmd.Command, distinctNodes(corr.NodePools, cmd.NodePoolNames...), corr.PrivateKey, corr.Vars)
		}

		if err != nil {
			corr.SetStatus(corral.StatusError)
			logrus.Error(err)
			break
		}

		// collect new nodes
		var nodes []corral.Node
		for name, np := range corr.NodePools {
			for _, n := range np {
				if _, ok := knownAddresses[n.Address]; !ok {
					n.OverlayRoot = pkg.Overlay[name]
					nodes = append(nodes, n)
					knownAddresses[n.Address] = knownAddresses[n.Address]
				}
			}
		}

		// copy package files to new nodes
		err = copyPackageFiles(nodes, corr.PrivateKey, pkg)
		if err != nil {
			corr.SetStatus(corral.StatusError)
			logrus.Error("failed to copy package files: ", err)
			return
		}

		_ = corr.Save()

	}

	if corr.Status == corral.StatusError {
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

	logrus.Info("done!")
	corr.SetStatus(corral.StatusReady)
}

func newShell(node corral.Node, privateKey string, vs vars.VarSet) (*shell.Shell, error) {
	sh := &shell.Shell{
		Node:       node,
		PrivateKey: []byte(privateKey),
		Vars:       vs,
	}

	err := sh.Connect()
	if err != nil {
		_ = sh.Close()
		return nil, err
	}

	return sh, nil
}

func copyPackageFiles(nodes []corral.Node, privateKey string, pkg _package.Package) error {
	var wg errgroup.Group
	for _, n := range nodes {
		n := n
		wg.Go(func() error {
			var sh *shell.Shell
			var err error

			err = wait.Poll(time.Second, 2*time.Minute, func() (done bool, err error) {
				sh, err = newShell(n, privateKey, nil)
				return err == nil, nil
			})
			if err != nil {
				return err
			}

			err = sh.UploadPackageFiles(pkg)
			_ = sh.Close()

			return err
		})
	}

	return wg.Wait()
}

func executeShellCommand(command string, nodes []corral.Node, privateKey string, vs vars.VarSet) error {
	var mu sync.Mutex
	var wg errgroup.Group
	for _, n := range nodes {
		n := n

		lvs := make(vars.VarSet)
		for k, v := range vs {
			lvs[k] = v
		}

		wg.Go(func() error {
			var sh *shell.Shell
			var err error

			err = wait.Poll(time.Second, 2*time.Minute, func() (done bool, err error) {
				sh, err = newShell(n, privateKey, lvs)
				return err == nil, nil
			})
			if err != nil {
				return err
			}

			sh.Run(command)

			err = sh.Close()
			if err != nil {
				return err
			}

			mu.Lock()
			for k, v := range sh.Vars {
				vs[k] = v
			}
			mu.Unlock()

			return nil
		})
	}

	return wg.Wait()
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

func distinctNodes(nodePools map[string][]corral.Node, poolNames ...string) (nodes []corral.Node) {
	seen := map[string]struct{}{}

	for _, name := range poolNames {
		if np := nodePools[name]; np != nil {
			for _, n := range np {
				if _, ok := seen[n.Address]; !ok {
					seen[n.Address] = seen[n.Address]
					nodes = append(nodes, n)
				}
			}
		}
	}

	return nodes
}
