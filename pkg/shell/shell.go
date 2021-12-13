package shell

import (
    "bufio"
    "fmt"
    "io"
    "io/fs"
    "net"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/pkg/sftp"
    "github.com/rancherlabs/corral/pkg/corral"
    _package "github.com/rancherlabs/corral/pkg/package"
    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/ssh"
)

const (
    PackageScriptDestination = "/opt/corral"
    connectionTimeout        = 5 * time.Second

    corralSetVarCommand = "corral_set"
)

type Shell struct {
    Node       corral.Node
    PrivateKey []byte
    Vars       map[string]string
    Verbose    bool

    stdin  chan []byte
    stdout chan []byte
    stderr chan []byte

    sftpClient    *sftp.Client
    bastionClient *ssh.Client
    client        *ssh.Client
    connection    net.Conn
    session       *ssh.Session
}

func (s *Shell) Connect() error {
    if len(strings.Split(s.Node.Address, ":")) == 1 {
        s.Node.Address += ":22"
    }

    if s.Node.BastionAddress != "" && len(strings.Split(s.Node.BastionAddress, ":")) == 1 {
        s.Node.BastionAddress += ":22"
    }

    signer, err := ssh.ParsePrivateKey(s.PrivateKey)
    if err != nil {
        return err
    }

    sshConfig := ssh.ClientConfig{
        User:    s.Node.User,
        Timeout: connectionTimeout,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
            return nil
        },
    }

    // establish a connection to the server
    if s.Node.BastionAddress != "" {
        s.bastionClient, err = ssh.Dial("tcp", s.Node.BastionAddress, &sshConfig)
        if err != nil {
            return err
        }

        s.connection, err = s.bastionClient.Dial("tcp", s.Node.Address)
        if err != nil {
            return err
        }
    } else {
        s.connection, err = net.DialTimeout("tcp", s.Node.Address, connectionTimeout)
        if err != nil {
            return err
        }
    }

    // upgrade connection to ssh connection
    sshConn, cc, cr, err := ssh.NewClientConn(s.connection, s.Node.Address, &sshConfig)
    if err != nil {
        return err
    }

    // create ssh client
    s.client = ssh.NewClient(sshConn, cc, cr)

    // connect sftp client
    s.sftpClient, err = sftp.NewClient(s.client)
    if err != nil {
        return err
    }

    // test sftp connection
    _, err = s.sftpClient.Stat("/")
    if err != nil {
        return err
    }

    // start a user shell
    err = s.startSession()
    if err != nil {
        return err
    }

    return nil
}

func (s *Shell) UploadPackageFiles(pkg _package.Package) error {
    src := pkg.ScriptPath() + string(os.PathSeparator)
    return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        dir, fn := filepath.Split(strings.TrimPrefix(path, src))

        err = s.sftpClient.MkdirAll(strings.ReplaceAll(filepath.Join(PackageScriptDestination, dir), string(os.PathSeparator), "/"))
        if err != nil {
            return err
        }

        in, err := os.Open(path)
        defer func() { _ = in.Close() }()
        if err != nil {
            return err
        }

        out, err := s.sftpClient.Create(strings.ReplaceAll(filepath.Join(PackageScriptDestination, dir, fn), string(os.PathSeparator), "/"))
        defer func() { _ = out.Close() }()
        if err != nil {
            return err
        }

        err = out.Chmod(0o700)
        if err != nil {
            return err
        }

        logrus.Debugf("copying %s to [%s]:%s", filepath.Join(PackageScriptDestination, dir), s.Node.Name, filepath.Join(PackageScriptDestination, dir, fn))

        _, err = io.Copy(out, in)
        if err != nil {
            return err
        }

        return nil
    })
}

func (s *Shell) Run(c string) {
    s.stdin <- []byte(c)
    s.stdin <- []byte("\n")
}

func (s *Shell) Close() error {
    var err error

    if s.stdin != nil {
        s.Run("exit")
        go func() { close(s.stdin) }()
    }
    if s.stdout != nil {
        go func() { close(s.stdout) }()
    }
    if s.stderr != nil {
        go func() { close(s.stderr) }()
    }

    if s.session != nil {
        _ = s.session.Wait()
    }

    if s.sftpClient != nil {
        _ = s.sftpClient.Close()
    }

    if s.connection != nil {
        _ = s.connection.Close()
    }

    if s.bastionClient != nil {
        _ = s.bastionClient.Close()
    }

    return err
}

func (s *Shell) startSession() (err error) {
    s.session, err = s.client.NewSession()
    if err != nil {
        return err
    }

    go s.connectStdin()
    go s.connectStdout()

    if s.Verbose {
        go s.connectStdin()
    }

    err = s.session.Shell()
    if err != nil {
        return err
    }

    for k, v := range s.Vars {
        s.Run(fmt.Sprintf("export CORRAL_%s=\"%s\"\n", k, v))
    }

    return nil
}

func (s *Shell) connectStdin() {
    if s.stdin != nil {
        return
    }

    s.stdin = make(chan []byte, 0)
    stdin, _ := s.session.StdinPipe()

    for {
        select {
        case d := <-s.stdin:
            _, _ = stdin.Write(d)
        default:
        }
    }
}

func (s *Shell) connectStdout() {
    if s.stdout != nil {
        return
    }

    s.stdout = make(chan []byte, 0)
    stdout, _ := s.session.StdoutPipe()

    scanner := bufio.NewScanner(stdout)

    for scanner.Scan() {
        text := scanner.Text()

        if strings.HasPrefix(text, corralSetVarCommand) {
            vs := strings.TrimPrefix(text, corralSetVarCommand)
            vs = strings.Trim(vs, " \t")

            k, v := corral.ToVar(vs)
            if k == "" {
                logrus.Warnf("failed to parse corral command: %s", text)
                continue
            }

            s.Vars[k] = v
        }

        if s.Verbose {
            fmt.Printf("[%s]: %s\n", s.Node.Address, scanner.Text())
        }
    }
}

func (s *Shell) connectStderr() {
    if s.stderr != nil {
        return
    }

    s.stderr = make(chan []byte, 10)
    stderr, _ := s.session.StderrPipe()

    scanner := bufio.NewScanner(stderr)

    for scanner.Scan() {
        fmt.Printf("[%s]: %s\n", s.Node.Address, scanner.Text())
    }
}
