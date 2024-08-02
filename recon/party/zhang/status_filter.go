package zhang

import (
	"context"

	"github.com/ivxivx/go-recon/recon/domain"
	txn "github.com/ivxivx/go-recon/recon/transaction"
)

type StatusFilter struct {
	validStatuses []string
}

func NewStatusFilter() *StatusFilter {
	filter := &StatusFilter{}

	filter.WithValidStatuses(StatusCompleted, StatusFailed)

	return filter
}

func (f *StatusFilter) WithValidStatuses(validStatuses ...string) *StatusFilter {
	f.validStatuses = validStatuses

	return f
}

var _ txn.Filter = (*StatusFilter)(nil)

func (f *StatusFilter) Filter(_ context.Context, transaction domain.Transaction) (bool, error) {
	ptx, cok := transaction.(*Transaction)
	if !cok {
		// do not handle this type of transaction
		return true, nil
	}

	for _, status := range f.validStatuses {
		if ptx.Status == status {
			return true, nil
		}
	}

	return false, nil
}
