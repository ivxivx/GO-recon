package transformer

import (
	"context"
)

type RecordExtractor interface {
	GetInputDataType() any
	Extract(ctx context.Context, data any) ([]any, error)
}
