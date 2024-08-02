package batch

import "context"

type Writer interface {
	OpenCloser

	Write(ctx context.Context, record any) error
}
