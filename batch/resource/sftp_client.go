package resource

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/ivxivx/go-recon/batch"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ISftpClient interface {
	Open(ctx context.Context) error
	Close(ctx context.Context) error
	OpenFile(path string, flag int) (*sftp.File, error)
}

type SSHServer struct {
	address string
	config  *ssh.ClientConfig
}

func NewSSHServer(
	address string,
	config *ssh.ClientConfig,
) *SSHServer {
	return &SSHServer{
		address: address,
		config:  config,
	}
}

// resource on sftp server
type SftpClient struct {
	logger     *slog.Logger
	destServer *SSHServer

	openOnce  sync.Once
	closeOnce sync.Once

	delegate *sftp.Client
}

var _ ISftpClient = (*SftpClient)(nil)

func NewSftpClient(logger *slog.Logger, destServer *SSHServer) *SftpClient {
	return &SftpClient{
		logger:     logger,
		destServer: destServer,
	}
}

func NewAuthMethodFromPrivateKeyFile(privateKeyFile string) (ssh.AuthMethod, error) {
	keyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file %s: %w", privateKeyFile, err)
	}

	return NewAuthMethodFromPrivateKey(keyData)
}

func NewAuthMethodFromPrivateKey(privateKey []byte) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not create signer from private key: %w", err)
	}

	authMethod := ssh.PublicKeys(signer)

	return authMethod, nil
}

func (c *SftpClient) Open(_ context.Context) error {
	if c.delegate != nil {
		c.logger.Info("connection is already opened", slog.String("address", c.destServer.address))

		return nil
	}

	var errR error

	c.openOnce.Do(func() {
		destServerClient, err := ssh.Dial("tcp", c.destServer.address, c.destServer.config)
		if err != nil {
			errR = &batch.ConnectionError{Operation: batch.ConnOpen, Address: c.destServer.address, Err: err}

			return
		}

		sftpClient, err := sftp.NewClient(destServerClient)
		if err != nil {
			errR = &batch.ConnectionError{Operation: batch.ConnOpen, Address: c.destServer.address, Err: err}

			return
		}

		c.delegate = sftpClient
	})

	return errR
}

func (c *SftpClient) Close(_ context.Context) error {
	if c.delegate == nil {
		c.logger.Info("connection is not opened, skip close", slog.String("address", c.destServer.address))

		return nil
	}

	var errR error

	c.closeOnce.Do(func() {
		if err := c.delegate.Close(); err != nil {
			errR = &batch.ConnectionError{Operation: batch.ConnClose, Address: c.destServer.address, Err: err}
		}

		c.delegate = nil
	})

	return errR
}

func (c *SftpClient) OpenFile(path string, flag int) (*sftp.File, error) {
	file, err := c.delegate.OpenFile(path, flag)
	if err != nil {
		return nil, &batch.IoError{Operation: batch.IoOpen, Resource: path, Err: err}
	}

	return file, nil
}
