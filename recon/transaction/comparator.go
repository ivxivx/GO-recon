package transaction

import (
	"context"

	"github.com/ivxivx/go-recon/recon/domain"
)

type Comparator interface {
	Compare(ctx context.Context, partyTransaction1, partyTransaction2 domain.Transaction) ([]*domain.TxReconItem, error)
}
