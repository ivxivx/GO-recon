package zhang

import (
	"context"
	"log/slog"

	"github.com/shopspring/decimal"

	"github.com/ivxivx/go-recon/recon"
	"github.com/ivxivx/go-recon/recon/domain"
	"github.com/ivxivx/go-recon/recon/party/wang"
	"github.com/ivxivx/go-recon/recon/transaction"
)

const (
	maxCount = 100
)

type Comparator struct {
	Logger *slog.Logger
}

var _ transaction.Comparator = &Comparator{}

func (cpr *Comparator) Compare(
	_ context.Context,
	partyTransaction1, partyTransaction2 domain.Transaction,
) ([]*domain.TxReconItem, error) {
	var tx1 *wang.Transaction

	if partyTransaction1 != nil {
		temp, cok := partyTransaction1.(*wang.Transaction)
		if !cok {
			return nil, &recon.UnexpectedTypeError{FromType: partyTransaction1, ToType: wang.Transaction{}}
		}

		tx1 = temp
	}

	var tx2 *Transaction

	if partyTransaction2 != nil {
		temp, cok := partyTransaction2.(*Transaction)
		if !cok {
			return nil, &recon.UnexpectedTypeError{FromType: partyTransaction2, ToType: Transaction{}}
		}

		tx2 = temp
	}

	reconItems := make([]*domain.TxReconItem, 0, maxCount)

	reconItem := cpr.compareStatus(tx1, tx2)

	reconItems = append(reconItems, reconItem)

	reconItem = cpr.compareCurrency(tx1, tx2)

	reconItems = append(reconItems, reconItem)

	reconItem = cpr.compareAmount(tx1, tx2)

	reconItems = append(reconItems, reconItem)

	return reconItems, nil
}

func (cpr *Comparator) cmpStatus(partyValue1, partyValue2 string) bool {
	switch partyValue1 {
	case wang.StatusCompleted:
		return partyValue2 == StatusCompleted
	case wang.StatusDeclined:
		return partyValue2 == StatusFailed
	default:
		return false
	}
}

func (cpr *Comparator) compareStatus(
	partyTransaction1 *wang.Transaction,
	partyTransaction2 *Transaction,
) *domain.TxReconItem {
	var partyValue1, partyValue2 *string

	if partyTransaction1 != nil {
		temp := partyTransaction1.Status
		partyValue1 = &temp
	}

	if partyTransaction2 != nil {
		temp := partyTransaction2.Status
		partyValue2 = &temp
	}

	var matched bool

	if partyTransaction1 == nil || partyTransaction2 == nil {
		matched = false
	} else {
		matched = cpr.cmpStatus(*partyValue1, *partyValue2)
	}

	return &domain.TxReconItem{
		Type:        string(domain.ItemTypeStatus),
		Key:         string(reconItemKeyStatus),
		PartyValue1: partyValue1,
		PartyValue2: partyValue2,
		Matched:     matched,
	}
}

func (cpr *Comparator) compareCurrency(
	partyTransaction1 *wang.Transaction,
	partyTransaction2 *Transaction,
) *domain.TxReconItem {
	var partyValue1, partyValue2 *string

	if partyTransaction1 != nil {
		temp := partyTransaction1.ReceivingCurrency
		partyValue1 = &temp
	}

	if partyTransaction2 != nil {
		temp := partyTransaction2.LocalCurrency
		partyValue2 = &temp
	}

	var matched bool
	if partyTransaction1 == nil || partyTransaction2 == nil {
		matched = false
	} else {
		matched = *partyValue1 == *partyValue2
	}

	return &domain.TxReconItem{
		Type:        string(domain.ItemTypeCurrency),
		Key:         string(reconItemKeyCurrency),
		PartyValue1: partyValue1,
		PartyValue2: partyValue2,
		Matched:     matched,
	}
}

func (cpr *Comparator) compareAmount(
	partyTransaction1 *wang.Transaction,
	partyTransaction2 *Transaction,
) *domain.TxReconItem {
	var partyValue1, partyValue2 *string

	var partyAmount1, partyAmount2, difference *decimal.Decimal

	if partyTransaction1 != nil {
		temp1 := partyTransaction1.ReceivingAmount
		partyAmount1 = &temp1

		temp2 := partyAmount1.String()
		partyValue1 = &temp2
	}

	if partyTransaction2 != nil {
		amount2, err := decimal.NewFromString(partyTransaction2.LocalAmount)
		if err != nil {
			cpr.Logger.Warn("failed to parse payout amount from partyTransaction2", slog.Any("error", err))
		}

		partyAmount2 = &amount2

		temp := partyAmount2.String()
		partyValue2 = &temp
	}

	var matched bool
	if partyAmount1 == nil || partyAmount2 == nil {
		matched = false
	} else {
		temp := partyAmount2.Sub(*partyAmount1)
		difference = &temp

		matched = partyAmount1.Cmp(*partyAmount2) == 0
	}

	return &domain.TxReconItem{
		Type:        string(domain.ItemTypeAmount),
		Key:         string(reconItemKeyAmount),
		PartyValue1: partyValue1,
		PartyValue2: partyValue2,
		Matched:     matched,
		Difference:  difference,
	}
}
