package transaction

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/ivxivx/go-recon/recon"
	"github.com/ivxivx/go-recon/recon/domain"
)

type Reconciler[T1, T2 domain.Transaction] struct {
	logger             *slog.Logger
	party1ID           string
	party2ID           string
	party1TxCollection Collection
	party2TxCollection Collection
	filter             Filter
	comparator         Comparator
}

func NewReconciler[T1, T2 domain.Transaction](
	logger *slog.Logger,
	party1ID, party2ID string,
	party1TxCollection, party2TxCollection Collection,
	comparator Comparator,
) *Reconciler[T1, T2] {
	return &Reconciler[T1, T2]{
		logger:             logger,
		party1ID:           party1ID,
		party2ID:           party2ID,
		party1TxCollection: party1TxCollection,
		party2TxCollection: party2TxCollection,
		comparator:         comparator,
	}
}

func (rc *Reconciler[T1, T2]) WithFilter(filter Filter) *Reconciler[T1, T2] {
	rc.filter = filter

	return rc
}

type ReconResult struct {
	// matching key -> result
	BothParties map[string]*domain.TxReconResult
	// party1 transaction id -> result
	Party1Only map[string]*domain.TxReconResult
	// party2 transaction id -> result
	Party2Only map[string]*domain.TxReconResult
}

type ReconResultCount struct {
	Matched    int
	Mismatched int
	Party1Only int
	Party2Only int
}

func (rr *ReconResult) GetCount() ReconResultCount {
	var matchedCount, mismatchedCount int

	for _, txReconResult := range rr.BothParties {
		if txReconResult.ResultType == recon.ResultMatched {
			matchedCount++
		} else {
			mismatchedCount++
		}
	}

	return ReconResultCount{
		Matched:    matchedCount,
		Mismatched: mismatchedCount,
		Party1Only: len(rr.Party1Only),
		Party2Only: len(rr.Party2Only),
	}
}

func (rc *Reconciler[T1, T2]) Process(ctx context.Context) (*ReconResult, error) {
	err := rc.party1TxCollection.Open(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errC := rc.party1TxCollection.Close(ctx); errC != nil {
			rc.logger.Warn("failed to close collection1", slog.Any("error", errC))

			return
		}
	}()

	err = rc.party2TxCollection.Open(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errC := rc.party2TxCollection.Close(ctx); errC != nil {
			rc.logger.Warn("failed to close collection2", slog.Any("error", errC))

			return
		}
	}()

	reconResult := &ReconResult{
		BothParties: make(map[string]*domain.TxReconResult),
		Party1Only:  make(map[string]*domain.TxReconResult),
		Party2Only:  make(map[string]*domain.TxReconResult),
	}

	err = rc.compareParty2AgainstParty1(ctx, reconResult)
	if err != nil {
		return nil, err
	}

	err = rc.compareParty1AgainstParty2(ctx, reconResult)
	if err != nil {
		return nil, err
	}

	return reconResult, nil
}

func (rc *Reconciler[T1, T2]) compareParty2AgainstParty1(
	ctx context.Context,
	reconResult *ReconResult,
) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
			err := rc.compare(ctx, reconResult, false)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break loop
				}

				return err
			}
		}
	}

	return nil
}

func (rc *Reconciler[T1, T2]) compareParty1AgainstParty2(
	ctx context.Context,
	reconResult *ReconResult,
) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
			err := rc.compare(ctx, reconResult, true)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break loop
				}

				return err
			}
		}
	}

	return nil
}

func (rc *Reconciler[T1, T2]) compare(
	ctx context.Context,
	reconResult *ReconResult,
	isParty1 bool,
) error {
	var partyTransaction1 domain.Transaction

	var notFoundResultType string

	var partyCollection1, partyCollection2 Collection

	if isParty1 {
		partyCollection1 = rc.party1TxCollection
		partyCollection2 = rc.party2TxCollection

		notFoundResultType = recon.ResultParty1Only

		var temp T1
		partyTransaction1 = temp
	} else {
		partyCollection1 = rc.party2TxCollection
		partyCollection2 = rc.party1TxCollection

		notFoundResultType = recon.ResultParty2Only

		var temp T2
		partyTransaction1 = temp
	}

	err := partyCollection1.Read(ctx, &partyTransaction1)
	if err != nil {
		return err
	}

	if rc.filter != nil {
		pass, errF := rc.filter.Filter(ctx, partyTransaction1)
		if errF != nil {
			return errF
		}

		if !pass {
			// do not process this transaction
			return nil
		}
	}

	matchingKey := partyTransaction1.GetMatchingKey()

	partyTransaction2, found := partyCollection2.Find(ctx, matchingKey)

	var party1Transaction, party2Transaction domain.Transaction

	if isParty1 {
		party1Transaction = partyTransaction1
		party2Transaction = partyTransaction2
	} else {
		party1Transaction = partyTransaction2
		party2Transaction = partyTransaction1
	}

	txReconItems, err := rc.comparator.Compare(ctx, party1Transaction, party2Transaction)
	if err != nil {
		return err
	}

	var resultType string

	if found {
		resultType = rc.deriveResultType(txReconItems)
	} else {
		resultType = notFoundResultType
	}

	txReconResult, err := rc.buildResult(
		matchingKey,
		party1Transaction,
		party2Transaction,
		resultType,
		txReconItems,
	)
	if err != nil {
		return err
	}

	if found {
		reconResult.BothParties[matchingKey] = txReconResult
	} else {
		if isParty1 {
			reconResult.Party1Only[partyTransaction1.GetID()] = txReconResult
		} else {
			reconResult.Party2Only[partyTransaction1.GetID()] = txReconResult
		}
	}

	return nil
}

func (rc *Reconciler[T1, T2]) deriveResultType(txReconItems []*domain.TxReconItem) string {
	var mismatchedType string

	for _, reconItem := range txReconItems {
		if reconItem.Matched {
			continue
		}

		if mismatchedType == "" {
			mismatchedType = reconItem.Type
		} else if mismatchedType != reconItem.Type {
			return recon.ResultMismatched
		}
	}

	if mismatchedType == "" {
		return recon.ResultMatched
	}

	return mismatchedType
}

func (rc *Reconciler[T1, T2]) buildResult(
	matchingKey string,
	partyTransaction1, partyTransaction2 domain.Transaction,
	resultType string,
	reconItems []*domain.TxReconItem,
) (*domain.TxReconResult, error) {
	var txID1, txID2 *string

	var timestamp time.Time

	var txType string

	switch resultType {
	case recon.ResultParty1Only:
		temp1 := partyTransaction1.GetID()
		txID1 = &temp1

		temp2 := partyTransaction1.GetExternalID()
		txID2 = temp2

		timestamp = partyTransaction1.GetTimestamp()
		txType = partyTransaction1.GetType()
	case recon.ResultParty2Only:
		temp1 := partyTransaction2.GetExternalID()
		txID1 = temp1

		temp2 := partyTransaction2.GetID()
		txID2 = &temp2

		timestamp = partyTransaction2.GetTimestamp()
		txType = partyTransaction2.GetType()
	default:
		temp1 := partyTransaction1.GetID()
		txID1 = &temp1

		temp2 := partyTransaction2.GetID()
		txID2 = &temp2

		timestamp = partyTransaction1.GetTimestamp()
		txType = partyTransaction1.GetType()
	}

	return &domain.TxReconResult{
		MatchingKey:          matchingKey,
		ResultType:           resultType,
		TransactionTimestamp: timestamp,
		TransactionType:      txType,
		PartyID1:             rc.party1ID,
		PartyID2:             rc.party2ID,
		PartyTransactionID1:  txID1,
		PartyTransactionID2:  txID2,
		Items:                reconItems,
	}, nil
}
