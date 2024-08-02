package filter

import (
	"context"
	"time"

	"github.com/ivxivx/go-recon/recon/domain"
	txn "github.com/ivxivx/go-recon/recon/transaction"
)

type TimestampFilter struct {
	startTimestamp *time.Time
	endTimestamp   *time.Time
}

func NewTimestampFilter(startTimestamp, endTimestamp *time.Time) *TimestampFilter {
	return &TimestampFilter{
		startTimestamp: startTimestamp,
		endTimestamp:   endTimestamp,
	}
}

var _ txn.Filter = (*TimestampFilter)(nil)

func (f *TimestampFilter) Filter(_ context.Context, transaction domain.Transaction) (bool, error) {
	timestamp := transaction.GetTimestamp()

	if f.startTimestamp != nil && timestamp.Before(*f.startTimestamp) {
		return false, nil
	}

	if f.endTimestamp != nil && timestamp.After(*f.endTimestamp) {
		return false, nil
	}

	return true, nil
}
