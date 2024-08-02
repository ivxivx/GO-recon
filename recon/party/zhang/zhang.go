package zhang

import (
	"time"

	"github.com/ivxivx/go-recon/recon/domain"
)

const (
	StatusCompleted string = "Completed"
	StatusFailed    string = "Canceled"

	reconItemKeyStatus   domain.ItemKey = "status"
	reconItemKeyCurrency domain.ItemKey = "currency"
	reconItemKeyAmount   domain.ItemKey = "amount"
)

type Transaction struct {
	CreationDate          time.Time `csv:"CREATION_DATE"`
	ExternalTransactionID string    `csv:"EXTERNAL_TRANSACTION_ID"`
	TransactionID         string    `csv:"TRANSACTION_ID"`
	LocalCurrency         string    `csv:"LOCAL_CURRENCY"`
	LocalAmount           string    `csv:"LOCAL_AMOUNT"`
	Status                string    `csv:"STATUS"`
}

func (t *Transaction) GetMatchingKey() string {
	return t.ExternalTransactionID
}

func (t *Transaction) GetID() string {
	return t.TransactionID
}

func (t *Transaction) GetExternalID() *string {
	temp := t.ExternalTransactionID

	return &temp
}

func (t *Transaction) GetType() string {
	return "payout"
}

func (t *Transaction) GetTimestamp() time.Time {
	return t.CreationDate
}

var _ domain.Transaction = (*Transaction)(nil)
