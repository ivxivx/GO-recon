package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ItemType string

const (
	ItemTypeStatus   ItemType = "status"
	ItemTypeAmount   ItemType = "amount"
	ItemTypeCurrency ItemType = "currency"
)

type ItemKey string

type TxReconItem struct {
	ID          uuid.UUID        `json:"id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ResultID    uuid.UUID        `json:"result_id"`
	Matched     bool             `json:"matched"`
	Type        string           `json:"type"`
	Key         string           `json:"key"`
	PartyValue1 *string          `json:"party_value1"`
	PartyValue2 *string          `json:"party_value2"`
	Difference  *decimal.Decimal `json:"difference"` // party2_value - party1_value, only for decimal values
}

func NewTxReconItemID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
