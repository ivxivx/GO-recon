package wang

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ivxivx/go-recon/batch/transformer"
	"github.com/ivxivx/go-recon/recon"
	"github.com/ivxivx/go-recon/recon/domain"
)

const (
	StatusCompleted string = "completed"
	StatusDeclined  string = "declined"
)

type Transaction struct {
	ID                    string          `json:"id"`
	CreatedAt             time.Time       `json:"created_at"`
	Status                string          `json:"status"`
	ReceivingAmount       decimal.Decimal `json:"receiving_amount"`
	ReceivingCurrency     string          `json:"receiving_currency"`
	ProviderTransactionID *string         `json:"provider_transaction_id"`
}

func (t *Transaction) GetMatchingKey() string {
	return t.ID
}

func (t *Transaction) GetID() string {
	return t.ID
}

func (t *Transaction) GetExternalID() *string {
	if t.ProviderTransactionID == nil {
		return nil
	}

	temp := *t.ProviderTransactionID

	return &temp
}

func (t *Transaction) GetType() string {
	return "payout"
}

func (t *Transaction) GetTimestamp() time.Time {
	return t.CreatedAt
}

var _ domain.Transaction = (*Transaction)(nil)

type GetTransfersReqeust struct {
	Party2ID string `json:"party2_id"`
}

type RecordExtractor struct{}

var _ transformer.RecordExtractor = (*RecordExtractor)(nil)

func (rt *RecordExtractor) GetInputDataType() any {
	return []*Transaction{}
}

func (rt *RecordExtractor) Extract(_ context.Context, data any) ([]any, error) {
	res, ok := data.([]*Transaction)
	if !ok {
		return nil, &recon.UnexpectedTypeError{FromType: data, ToType: []*Transaction{}}
	}

	return transformer.ConvertSlice(res), nil
}
