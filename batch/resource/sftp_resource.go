package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/pkg/sftp"

	"github.com/ivxivx/go-recon/batch"
)

// resource on sftp server
type SftpResource struct {
	logger   *slog.Logger
	client   ISftpClient
	filePath string
	flag     int // default is os.O_RDONLY

	openOnce  sync.Once
	closeOnce sync.Once

	file *sftp.File
}

func NewSftpResource(logger *slog.Logger, client ISftpClient, filePath string) *SftpResource {
	return &SftpResource{
		logger:   logger,
		client:   client,
		filePath: filePath,
		flag:     os.O_RDONLY,
	}
}

func (r *SftpResource) WithFlag(flag int) *SftpResource {
	r.flag = flag

	return r
}

var (
	_ batch.Resource = (*SftpResource)(nil)
	_ io.Reader      = (*SftpResource)(nil)
	_ io.Writer      = (*SftpResource)(nil)
)

func (r *SftpResource) GetID() string {
	return r.filePath
}

func (r *SftpResource) Open(ctx context.Context) error {
	if r.file != nil {
		r.logger.Warn("resource is already opened", slog.String("resource", r.filePath))

		return nil
	}

	var errR error

	r.openOnce.Do(func() {
		err := r.client.Open(ctx)
		if err != nil {
			errR = fmt.Errorf("could not open sftp client for %s: %w", r.filePath, err)

			return
		}

		file, err := r.client.OpenFile(r.filePath, r.flag)
		if err != nil {
			errR = fmt.Errorf("could not open file %s via sftp: %w", r.filePath, err)

			return
		}

		r.file = file
	})

	return errR
}

func (r *SftpResource) Close(ctx context.Context) error {
	if r.file == nil {
		r.logger.Info("resource is not opened, skip close", slog.String("resource", r.filePath))

		return nil
	}

	var errR error

	r.closeOnce.Do(func() {
		defer func() {
			r.file = nil
			r.client = nil
		}()

		if err := r.file.Close(); err != nil {
			errR = &batch.IoError{Operation: batch.IoClose, Resource: r.filePath, Err: err}
		}

		if r.client != nil {
			if errC := r.client.Close(ctx); errC != nil {
				r.logger.Warn("failed to close sftp client", slog.String("resource", r.filePath), slog.Any("error", errC))
			}
		}
	})

	return errR
}

func (r *SftpResource) Read(p []byte) (n int, err error) {
	if r.file == nil {
		return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.filePath, Err: ErrResourceNotOpened}
	}

	n, err = r.file.Read(p)

	if err == nil || errors.Is(err, io.EOF) {
		return n, err
	}

	return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.filePath, Err: err}
}

func (r *SftpResource) Write(p []byte) (n int, err error) {
	if r.file == nil {
		return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.filePath, Err: ErrResourceNotOpened}
	}

	n, err = r.file.Write(p)
	if err == nil {
		return n, nil
	}

	return 0, &batch.IoError{Operation: batch.IoWrite, Resource: r.filePath, Err: err}
}
