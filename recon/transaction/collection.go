package transaction

import (
	"context"

	"github.com/ivxivx/go-recon/batch"
	"github.com/ivxivx/go-recon/recon/domain"
)

type Collection interface {
	batch.Reader
	Find(ctx context.Context, matchingKey string) (domain.Transaction, bool)
}
