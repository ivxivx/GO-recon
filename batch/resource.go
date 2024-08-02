package batch

import (
	"context"
	"io"
)

type OpenCloser interface {
	Open(ctx context.Context) error
	Close(ctx context.Context) error
}

type Resource interface {
	OpenCloser
	GetID() string
}

type ResourceReader interface {
	Resource
	io.Reader
}

type ResourceWriter interface {
	Resource
	io.Writer
}
