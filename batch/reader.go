package batch

import "context"

type Reader interface {
	OpenCloser

	Read(ctx context.Context, record any) error
}
