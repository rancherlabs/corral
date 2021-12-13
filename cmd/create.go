package cmd

import (
    "bytes"
    "context"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/json"
    "encoding/pem"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/hashicorp/terraform-exec/tfexec"
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
            cfg = config.Load()

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

    cmd.Flags().String("config", "", "Load flags for this command from a file.")

    cmd.Flags().Bool("dry-run", false, "Display what resources this corral will create.")
    _ = cfgViper.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))

    cmd.Flags().StringArrayP("variable", "v", []string{}, "Set a variable to configure the package.")
    _ = cfgViper.BindPFlag("variable", cmd.Flags().Lookup("variable"))

    cmd.Flags().String("package", "", "Set a variable to configure the package.")
    _ = cfgViper.BindPFlag("package", cmd.Flags().Lookup("package"))

    return cmd
}

func create(_ *cobra.Command, args []string) {
    source := cfgViper.GetString("package")
    if source == "" {
        if len(args) < 2 {
            logrus.Fatal("You must specify a package with the `-p` flag or as an argument.")
        }

        source = args[1]
    }

    var corr corral.Corral

    corr.RootPath = cfg.CorralPath(args[0])
    corr.Name = args[0]
    corr.NodePools = map[string][]corral.Node{}
    corr.Vars = map[string]string{}

    // ensure this corral is unique
    if corr.Exists() {
        logrus.Fatalf("corral [%s] already exists", corr.Name)
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
    for _, raw := range cfgViper.GetStringSlice("variable") {
        k, v := corral.ToVar(raw)
        if k == "" {
            logrus.Fatal("variables should be in the format <key>=<value>")
        }
        corr.Vars[k] = v
    }
    for k, v := range cfg.Vars { // copy the global vars for future reference
        corr.Vars[k] = v
    }

    // load the package
    pkg, err := _package.LoadPackage(args[1], cfg.PackageCachePath(), cfg.RegistryCredentialsFile())
    if err != nil {
        logrus.Fatalf("failed to load package: %s", err)
    }

    // start a new tf instance in our corral's terraform path
    _ = corr.Save()
    tf, _ := tfexec.NewTerraform(corr.TerraformPath(), cfg.AppPath("bin", "terraform"))

    defer func() {
        if corr.Status < corral.StatusReady {
            if corr.Status > corral.StatusNew {
                logrus.Info("cleaning up failed corral")
                err = tf.Destroy(context.Background())
                if err != nil {
                    logrus.Error("failed to cleanup failed corral: ", err)
                    corr.SetStatus(corral.StatusError)
                    return
                }
            }

            _ = corr.Delete()
        }
    }()

    if debug || cfgViper.GetBool("dry-run") {
        tf.SetStdout(os.Stdout)
    }

    logrus.Info("initializing terraform module")
    err = tf.Init(context.Background(),
        tfexec.Get(false),
        tfexec.Upgrade(false),
        tfexec.FromModule(pkg.TerraformModulePath()),
        tfexec.PluginDir(pkg.TerraformPluginPath()))
    if err != nil {
        logrus.Error(err)
        return
    }

    // write the tf vars file
    f, err := os.Create(filepath.Join(corr.TerraformPath(), "terraform.tfvars.json"))
    if err != nil {
        logrus.Error(err)
    }
    _ = json.NewEncoder(f).Encode(corr.Vars)
    defer func() { _ = f.Close() }()

    // if this is a dry run just plan and exit
    if cfgViper.GetBool("dry-run") {
        _, err = tf.Plan(context.Background())
        if err != nil {
            logrus.Error("error calling plan: ", err)
        }

        // because this is a dry run we can delete the corral
        _ = corr.Delete()

        return
    }

    // write the corral to disk
    corr.SetStatus(corral.StatusProvisioning)

    logrus.Info("applying terraform module")
    err = tf.Apply(context.Background())
    if err != nil {
        logrus.Error("failed to apply package terraform modules: ", err)
        corr.SetStatus(corral.StatusError)
        return
    }

    logrus.Info("reading terraform output")
    tfOutput, err := tf.Output(context.Background())
    if err != nil {
        logrus.Fatal("failed to load outputs from terraform: ", err)
        corr.SetStatus(corral.StatusError)
        return
    }

    for k, v := range tfOutput {
        if k == "corral_node_pools" {
            err = json.Unmarshal(v.Value, &corr.NodePools)
            if err != nil {
                corr.SetStatus(corral.StatusError)
                logrus.Error("failed to parse corral node pools: ", err)
            }

            continue
        }

        var val string
        _ = json.Unmarshal(v.Value, &val)

        corr.Vars[k] = val
    }

    var nodes []corral.Node
    for _, ns := range corr.NodePools {
        nodes = append(nodes, ns...)
    }

    logrus.Info("copying package files to nodes")
    err = copyPackageFiles(nodes, &corr, pkg)
    if err != nil {
        corr.SetStatus(corral.StatusError)
        logrus.Error("failed to copy package files: ", err)
        return
    }

    for _, pkgCmd := range pkg.Commands {
        logrus.Infof("running command [%s]", pkgCmd.Command)

        err := executeCommand(pkgCmd, &corr)
        if err != nil {
            corr.SetStatus(corral.StatusError)
            logrus.Error("failed to execute command: ", err)
            return
        }
    }

    corr.SetStatus(corral.StatusReady)
}

func newShell(node corral.Node, c *corral.Corral) (*shell.Shell, error) {
    sh := &shell.Shell{
        Node:       node,
        PrivateKey: []byte(c.PrivateKey),
        Vars:       c.Vars,
        Verbose:    debug,
    }

    err := sh.Connect()
    if err != nil {
        _ = sh.Close()
        return nil, err
    }

    return sh, nil
}

func copyPackageFiles(nodes []corral.Node, c *corral.Corral, pkg _package.Package) error {
    var wg errgroup.Group
    for _, n := range nodes {
        n := n
        c := c
        wg.Go(func() error {
            var sh *shell.Shell
            var err error

            err = wait.Poll(time.Second, 2*time.Minute, func() (done bool, err error) {
                sh, err = newShell(n, c)
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

func executeCommand(pkgCmd _package.Command, c *corral.Corral) error {
    nodes := map[string]corral.Node{}
    for _, np := range pkgCmd.NodePoolNames {
        // if the nodepool does not exist skip it
        if _, ok := c.NodePools[np]; !ok {
            continue
        }

        // accumulate unique nodes
        for _, n := range c.NodePools[np] {
            nodes[n.Name] = n
        }
    }

    var mu sync.Mutex
    var wg errgroup.Group
    for _, n := range nodes {
        n := n
        c := c
        wg.Go(func() error {
            var sh *shell.Shell
            var err error

            err = wait.Poll(time.Second, 2*time.Minute, func() (done bool, err error) {
                sh, err = newShell(n, c)
                return err == nil, nil
            })
            if err != nil {
                return err
            }

            sh.Run(pkgCmd.Command)

            err = sh.Close()
            if err != nil {
                return err
            }

            mu.Lock()
            for k, v := range sh.Vars {
                c.Vars[k] = v
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
