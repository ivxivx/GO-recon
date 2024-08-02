package transaction

import (
	"context"

	"github.com/ivxivx/go-recon/recon/domain"
)

type Filter interface {
	Filter(ctx context.Context, transaction domain.Transaction) (bool, error)
}
