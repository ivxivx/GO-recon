package collection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/ivxivx/go-recon/batch"
	"github.com/ivxivx/go-recon/recon/domain"
	"github.com/ivxivx/go-recon/recon/transaction"
)

const (
	maxCount = 1000
)

type InMemoryCollection[T domain.Transaction] struct {
	reader batch.Reader

	items   []T
	index   int
	itemMap map[string]T
}

func NewInMemoryCollection[T domain.Transaction](
	reader batch.Reader,
) *InMemoryCollection[T] {
	return &InMemoryCollection[T]{
		reader: reader,
	}
}

var _ transaction.Collection = (*InMemoryCollection[domain.Transaction])(nil)

func (col *InMemoryCollection[T]) Open(ctx context.Context) error {
	err := col.reader.Open(ctx)
	if err != nil {
		return err
	}

	items := make([]T, 0, maxCount)
	itemMap := make(map[string]T, maxCount)

loop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
			var item T

			err := col.reader.Read(ctx, &item)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break loop
				}

				return err
			}

			items = append(items, item)
			itemMap[item.GetMatchingKey()] = item
		}
	}

	col.items = items
	col.itemMap = itemMap

	return nil
}

func (col *InMemoryCollection[T]) Close(ctx context.Context) error {
	return col.reader.Close(ctx)
}

func (col *InMemoryCollection[T]) Read(_ context.Context, record any) error {
	if col.items == nil {
		return io.EOF
	}

	if col.index >= len(col.items) {
		return io.EOF
	}

	outValue := reflect.ValueOf(record).Elem()
	inValue := reflect.ValueOf(col.items[col.index])

	outValue.Set(inValue)

	col.index++

	return nil
}

func (col *InMemoryCollection[T]) Find(_ context.Context, matchingKey string) (domain.Transaction, bool) {
	item, found := col.itemMap[matchingKey]
	if !found {
		return nil, false
	}

	return item, true
}
