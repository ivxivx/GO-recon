package reader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"sync"

	"github.com/goccy/go-json"

	"github.com/ivxivx/go-recon/batch"
	"github.com/ivxivx/go-recon/batch/transformer"
)

type JSONReader struct {
	logger          *slog.Logger
	resource        batch.ResourceReader
	recordExtractor transformer.RecordExtractor

	openOnce  sync.Once
	closeOnce sync.Once
	readOnce  sync.Once

	records []any
	index   int
}

func NewJSONReader(
	logger *slog.Logger,
	resource batch.ResourceReader,
	recordExtractor transformer.RecordExtractor,
) *JSONReader {
	return &JSONReader{
		logger:          logger,
		resource:        resource,
		recordExtractor: recordExtractor,
	}
}

var _ batch.Reader = (*JSONReader)(nil)

func (r *JSONReader) Open(ctx context.Context) error {
	if r.resource == nil {
		return &batch.IllegalArgumentError{Name: "resource"}
	}

	var errR error

	r.openOnce.Do(func() {
		err := r.resource.Open(ctx)
		if err != nil {
			errR = fmt.Errorf("could not open resource: %w", err)
		}
	})

	return errR
}

func (r *JSONReader) Close(ctx context.Context) error {
	if r.resource == nil {
		r.logger.Warn("resource is not provided")

		return nil
	}

	var errR error

	r.closeOnce.Do(func() {
		if err := r.resource.Close(ctx); err != nil {
			errR = fmt.Errorf("could not close resource: %w", err)
		}

		r.resource = nil
	})

	return errR
}

func (r *JSONReader) Read(ctx context.Context, record any) error {
	if record == nil {
		return &batch.IllegalArgumentError{Name: "record"}
	}

	var errR error

	r.readOnce.Do(func() {
		rawData, err := io.ReadAll(r.resource)
		if err != nil {
			errR = &batch.IoError{Operation: batch.IoRead, Resource: r.resource.GetID(), Err: err}

			return
		}

		records, err := r.transform(ctx, rawData)
		if err != nil {
			errR = err

			return
		}

		r.records = records
		r.index = 0
	})

	if errR != nil {
		return errR
	}

	if r.records == nil {
		return io.EOF
	}

	if r.index >= len(r.records) {
		return io.EOF
	}

	outValue := reflect.ValueOf(record).Elem()
	inValue := reflect.ValueOf(r.records[r.index])
	outValue.Set(inValue)

	r.index++

	return nil
}

func (r *JSONReader) transform(ctx context.Context, raw []byte) ([]any, error) {
	dataType := r.recordExtractor.GetInputDataType()

	dataPtr := reflect.New(reflect.TypeOf(dataType)).Interface()

	// transform raw data to an object
	err := json.Unmarshal(raw, dataPtr)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal raw data: %w", err)
	}

	data := reflect.ValueOf(dataPtr).Elem().Interface()

	// transform the object to records
	records, err := r.recordExtractor.Extract(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("could not extract records: %w", err)
	}

	return records, nil
}
