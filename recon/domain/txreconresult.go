package domain

import (
	"time"

	"github.com/google/uuid"
)

type TxReconResult struct {
	ID                   uuid.UUID      `json:"id"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	MatchingKey          string         `json:"matching_key"`
	ResultType           string         `json:"result_type"`
	TransactionTimestamp time.Time      `json:"transaction_timestamp"`
	TransactionType      string         `json:"transaction_type"`
	PartyID1             string         `json:"party_id1"`
	PartyID2             string         `json:"party_id2"`
	PartyTransactionID1  *string        `json:"party_transaction_id1,omitempty"`
	PartyTransactionID2  *string        `json:"party_transaction_id2,omitempty"`
	Items                []*TxReconItem `json:"items,omitempty"`
}

func NewTxReconResultID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
