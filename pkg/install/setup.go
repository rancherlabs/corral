package install

import (
    "context"
    "os"

    tfversion "github.com/hashicorp/go-version"
    "github.com/hashicorp/hc-install/product"
    "github.com/hashicorp/hc-install/releases"
    "github.com/rancherlabs/corral/pkg/config"
    "github.com/rancherlabs/corral/pkg/version"
)

func Install(userID, userPublicKeyPath string, vars map[string]string) (err error) {
    cfg := &config.Config{
        UserID:            userID,
        UserPublicKeyPath: userPublicKeyPath,
        Version:           version.Version,
        Vars:              vars,
    }

    err = cfg.Save()
    if err != nil {
        return err
    }

    _ = os.MkdirAll(cfg.AppPath("bin"), 0o700)
    if err != nil {
        return err
    }

    tfInstaller := &releases.ExactVersion{
        Product:    product.Terraform,
        Version:    tfversion.Must(tfversion.NewVersion(version.TerraformVersion)),
        InstallDir: cfg.AppPath("bin"),
    }

    _, err = tfInstaller.Install(context.Background())
    if err != nil {
        return err
    }

    return nil
}
