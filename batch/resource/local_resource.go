package resource

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/ivxivx/go-recon/batch"
)

const (
	defaultFileMode = 0o644
)

type LocalResource struct {
	Logger   *slog.Logger
	FilePath string
	Flag     int
	FileMode fs.FileMode

	openOnce  sync.Once
	closeOnce sync.Once

	file *os.File
}

func NewLocalResource(
	logger *slog.Logger,
	filePath string,
) *LocalResource {
	return &LocalResource{
		Logger:   logger,
		FilePath: filePath,
		Flag:     os.O_RDONLY,
		FileMode: 0,
	}
}

func (r *LocalResource) WithFlag(flag int) *LocalResource {
	r.Flag = flag

	return r
}

func (r *LocalResource) WithWriteFlag() *LocalResource {
	r.Flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	return r
}

func (r *LocalResource) WithFileMode(fileMode fs.FileMode) *LocalResource {
	r.FileMode = fileMode

	return r
}

func (r *LocalResource) WithDefaultFileMode() *LocalResource {
	r.FileMode = defaultFileMode

	return r
}

var (
	_ batch.Resource = (*LocalResource)(nil)
	_ io.Reader      = (*LocalResource)(nil)
	_ io.Writer      = (*LocalResource)(nil)
)

func (r *LocalResource) GetID() string {
	return r.FilePath
}

func (r *LocalResource) Open(_ context.Context) error {
	if r.file != nil {
		// one resource may be shared by multiple readers/writers, so can be opened multiple times
		r.Logger.Warn("resource is already opened", slog.String("resource", r.FilePath))

		return nil
	}

	var errR error

	r.openOnce.Do(func() {
		file, err := os.OpenFile(r.FilePath, r.Flag, r.FileMode)
		if err != nil {
			errR = &batch.IoError{Operation: batch.IoOpen, Resource: r.FilePath, Err: err}

			return
		}

		r.file = file
	})

	return errR
}

func (r *LocalResource) Close(_ context.Context) error {
	if r.file == nil {
		r.Logger.Info("resource is not opened, skip close", slog.String("resource", r.FilePath))

		return nil
	}

	var errR error

	r.closeOnce.Do(func() {
		if err := r.file.Close(); err != nil {
			errR = &batch.IoError{Operation: batch.IoClose, Resource: r.FilePath, Err: err}
		}

		r.file = nil
	})

	return errR
}

func (r *LocalResource) Read(p []byte) (n int, err error) {
	if r.file == nil {
		return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.FilePath, Err: ErrResourceNotOpened}
	}

	n, err = r.file.Read(p)

	if err == nil || errors.Is(err, io.EOF) {
		return n, err
	}

	return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.FilePath, Err: err}
}

func (r *LocalResource) Write(p []byte) (n int, err error) {
	if r.file == nil {
		return 0, &batch.IoError{Operation: batch.IoWrite, Resource: r.FilePath, Err: ErrResourceNotOpened}
	}

	n, err = r.file.Write(p)
	if err == nil {
		return n, nil
	}

	return 0, &batch.IoError{Operation: batch.IoWrite, Resource: r.FilePath, Err: err}
}
