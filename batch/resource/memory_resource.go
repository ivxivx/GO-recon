package resource

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"

	"github.com/ivxivx/go-recon/batch"
)

type MemoryResource struct {
	logger     *slog.Logger
	identifier string

	openOnce sync.Once

	buf *bytes.Buffer
}

func NewMemoryResource(
	logger *slog.Logger,
	identifier string,
) *MemoryResource {
	return &MemoryResource{
		logger:     logger,
		identifier: identifier,
	}
}

var (
	_ batch.Resource = (*MemoryResource)(nil)
	_ io.Reader      = (*MemoryResource)(nil)
	_ io.Writer      = (*MemoryResource)(nil)
)

func (r *MemoryResource) GetData() []byte {
	if r.buf == nil {
		return []byte{}
	}

	return r.buf.Bytes()
}

func (r *MemoryResource) GetID() string {
	return r.identifier
}

func (r *MemoryResource) Open(_ context.Context) error {
	if r.buf != nil {
		r.logger.Warn("resource is already opened", slog.String("resource", r.identifier))

		return nil
	}

	r.openOnce.Do(func() {
		r.buf = bytes.NewBuffer([]byte{})
	})

	return nil
}

func (r *MemoryResource) Close(_ context.Context) error {
	if r.buf == nil {
		r.logger.Info("resource is not opened, skip close", slog.String("resource", r.identifier))

		return nil
	}

	r.buf = nil

	return nil
}

func (r *MemoryResource) Read(p []byte) (n int, err error) {
	if r.buf == nil {
		return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.identifier, Err: ErrResourceNotOpened}
	}

	n, err = r.buf.Read(p)

	if err == nil || errors.Is(err, io.EOF) {
		return n, err
	}

	return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.identifier, Err: err}
}

func (r *MemoryResource) Write(p []byte) (n int, err error) {
	if r.buf == nil {
		return 0, &batch.IoError{Operation: batch.IoWrite, Resource: r.identifier, Err: ErrResourceNotOpened}
	}

	n, err = r.buf.Write(p)
	if err == nil {
		return n, nil
	}

	return 0, &batch.IoError{Operation: batch.IoWrite, Resource: r.identifier, Err: err}
}
