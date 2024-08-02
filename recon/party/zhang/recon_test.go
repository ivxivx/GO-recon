package zhang

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ivxivx/go-recon/batch/reader"
	rs "github.com/ivxivx/go-recon/batch/resource"
	"github.com/ivxivx/go-recon/batch/transformer"
	"github.com/ivxivx/go-recon/recon"
	"github.com/ivxivx/go-recon/recon/domain"
	"github.com/ivxivx/go-recon/recon/party"
	"github.com/ivxivx/go-recon/recon/party/wang"
	"github.com/ivxivx/go-recon/recon/transaction"
	"github.com/ivxivx/go-recon/recon/transaction/collection"
)

func Test_Recon(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recordExtractor := &wang.RecordExtractor{}

	testCases := []struct {
		name          string
		transactions  []*wang.Transaction
		txReconResult *domain.TxReconResult
	}{
		{
			name: "matching",
			transactions: []*wang.Transaction{
				{
					ID:                    "3bd0a9ee-c4ee-402f-8f39-f80642455838",
					CreatedAt:             time.Date(2024, 5, 31, 13, 45, 22, 0, time.UTC),
					Status:                wang.StatusCompleted,
					ReceivingAmount:       decimal.RequireFromString("500.00"),
					ReceivingCurrency:     "COP",
				},
			},
			txReconResult: &domain.TxReconResult{
				MatchingKey:          "3bd0a9ee-c4ee-402f-8f39-f80642455838",
				ResultType:           recon.ResultMatched,
				TransactionTimestamp: time.Date(2024, 5, 31, 13, 45, 22, 0, time.UTC),
				TransactionType:      "payout",
				PartyID1:             string(party.Wang),
				PartyID2:             string(party.Zhang),
				PartyTransactionID1:  ptr("3bd0a9ee-c4ee-402f-8f39-f80642455838"),
				PartyTransactionID2:  ptr("403020377"),
				Items: []*domain.TxReconItem{
					{Type: string(domain.ItemTypeStatus), Key: string(reconItemKeyStatus), PartyValue1: ptr("completed"), PartyValue2: ptr("Completed"), Matched: true},
					{Type: string(domain.ItemTypeCurrency), Key: string(reconItemKeyCurrency), PartyValue1: ptr("COP"), PartyValue2: ptr("COP"), Matched: true},
					{Type: string(domain.ItemTypeAmount), Key: string(reconItemKeyAmount), PartyValue1: ptr("500"), PartyValue2: ptr("500"), Matched: true, Difference: ptr(decimal.RequireFromString("0.00"))},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
				}

				if r.URL.Path != "/transactions" {
					t.Fatalf("unexpected path")
				}

				res, err := json.Marshal(tc.transactions)
				if err != nil {
					t.Fatalf("failed to marshal response: %v", err)
				}

				if num, err := w.Write(res); err != nil {
					t.Fatalf("failed to write response: %v, num: %d", err, num)
				}
			}))

			defer server.Close()

			resource1 := rs.NewHTTPResource(slog.Default(), server.URL+"/transactions")

			reader1 := reader.NewJSONReader(slog.Default(), resource1, recordExtractor)

			collection1 := collection.NewInMemoryCollection[*wang.Transaction](reader1)

			timeTransformer := &transformer.TimeTransformer{
				InputFormat:  time.DateTime,
				OutputFormat: time.RFC3339,
			}

			resource2 := rs.NewLocalResource(slog.Default(), "./testdata/Report_20240801.csv")

			reader2 := reader.NewCsvReader(
				slog.Default(),
				resource2,
			).WithTransformers(
				map[string]transformer.FieldTransformer{
					"CREATION_DATE": timeTransformer,
					"UPDATED_DATE":  timeTransformer,
				},
			)

			collection2 := collection.NewInMemoryCollection[*Transaction](reader2)

			comparator := &Comparator{
				Logger: slog.Default(),
			}

			reconciler := transaction.NewReconciler[*wang.Transaction, *Transaction](
				slog.Default(),
				string(string(party.Wang)),
				string(string(party.Zhang)),
				collection1,
				collection2,
				comparator,
			)

			reconResult, err := reconciler.Process(ctx)
			if err != nil {
				t.Fatalf("failed to reconcile: %v", err)
			}

			t.Logf("reconResult: %v", reconResult)

			var txReconResult *domain.TxReconResult

			var found bool

			switch tc.txReconResult.ResultType {
			case recon.ResultParty1Only:
				txReconResult, found = reconResult.Party1Only[*tc.txReconResult.PartyTransactionID1]
			case recon.ResultParty2Only:
				txReconResult, found = reconResult.Party2Only[*tc.txReconResult.PartyTransactionID2]
			case recon.ResultMatched:
				fallthrough
			default:
				{
					txReconResult, found = reconResult.BothParties[tc.txReconResult.MatchingKey]
				}
			}

			if !found {
				t.Fatalf("recon record not found for ID %s", tc.txReconResult.MatchingKey)
			}

			txReconResult.ID = uuid.Nil

			for _, reconItem := range txReconResult.Items {
				reconItem.ID = uuid.Nil
				reconItem.ResultID = uuid.Nil
			}

			if !cmp.Equal(txReconResult, tc.txReconResult) {
				t.Fatalf("recon record not matching, expected: %v, got: %v", tc.txReconResult, txReconResult)
			}
		})
	}
}

func ptr[T any](s T) *T {
	return &s
}
