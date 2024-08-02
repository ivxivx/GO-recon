package filter

import (
	"context"

	"github.com/ivxivx/go-recon/recon/domain"
	txn "github.com/ivxivx/go-recon/recon/transaction"
)

type AllPassFilter struct {
	filters []txn.Filter
}

func NewAllPassFilter(filters ...txn.Filter) *AllPassFilter {
	return &AllPassFilter{
		filters: filters,
	}
}

var _ txn.Filter = (*AllPassFilter)(nil)

func (f *AllPassFilter) Filter(ctx context.Context, transaction domain.Transaction) (bool, error) {
	for _, filter := range f.filters {
		pass, err := filter.Filter(ctx, transaction)
		if err != nil {
			return false, err
		}

		// if any filter fails, then this filter fails
		if !pass {
			return false, nil
		}
	}

	return true, nil
}
