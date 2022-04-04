package shell

import (
	"bufio"
	"fmt"
	"github.com/rancherlabs/corral/pkg/vars"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	connectionTimeout = 5 * time.Second

	corralSetVarCommand     = "corral_set"
	corralLogMessageCommand = "corral_log"
)

type Shell struct {
	Node       corral.Node
	PrivateKey []byte
	Vars       vars.VarSet

	sftpClient    *sftp.Client
	bastionClient *ssh.Client
	client        *ssh.Client
	connection    net.Conn
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

	return nil
}

func (s *Shell) UploadPackageFiles(pkg _package.Package) error {
	src := pkg.OverlayPath()
	if len(s.Node.OverlayRoot) > 0 {
		src = filepath.Join(src, s.Node.OverlayRoot)
	}

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dest := path[len(src):]

		if dest == "" {
			return nil
		}

		if info.IsDir() {
			return s.sftpClient.MkdirAll(dest)
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()

		out, err := s.sftpClient.Create(dest)
		if err != nil {
			return err
		}
		defer func() { _ = out.Close() }()

		err = out.Chmod(0o700)
		if err != nil {
			return err
		}

		logrus.Debugf("copying %s to [%s]:%s", path, s.Node.Name, dest)

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Shell) Run(c string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}

	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		s.consumeStdout(stdout)
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		s.consumeStderr(stderr)
		wg.Done()
	}()

	var request string
	for k, v := range s.Vars {
		request += fmt.Sprintf("export CORRAL_%s=\"%s\"\n", k, v)
	}
	request += c

	err = session.Run(fmt.Sprintf(request))
	wg.Wait()

	return err
}

func (s *Shell) Close() {
	if s.sftpClient != nil {
		_ = s.sftpClient.Close()
	}

	if s.connection != nil {
		_ = s.connection.Close()
	}

	if s.bastionClient != nil {
		_ = s.bastionClient.Close()
	}
}

func (s *Shell) consumeStdout(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, corralSetVarCommand) {
			cmd := strings.TrimPrefix(text, corralSetVarCommand)
			cmd = strings.Trim(cmd, " \t")

			k, v := vars.ToVar(cmd)
			if k == "" {
				logrus.Warnf("failed to parse corral command: %s", text)

				continue
			}

			s.Vars[k] = v
		} else if strings.HasPrefix(text, corralLogMessageCommand) {
			vs := strings.TrimPrefix(text, corralLogMessageCommand)
			vs = strings.Trim(vs, " \t")

			logrus.Info(vs)
		}

		logrus.Debugf("[%s]: %s", s.Node.Name, text)
	}
}

func (s *Shell) consumeStderr(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		logrus.Debugf("[%s]: %s", s.Node.Name, scanner.Text())
	}
}
