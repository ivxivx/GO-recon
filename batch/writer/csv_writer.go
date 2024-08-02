package writer

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ivxivx/go-recon/batch"
	"github.com/jszwec/csvutil"
)

type CsvWriter struct {
	logger     *slog.Logger
	resource   batch.ResourceWriter
	fieldNames []string

	openOnce  sync.Once
	closeOnce sync.Once

	encoded bool
	writer  *csv.Writer
	encoder *csvutil.Encoder
}

func NewCsvWriter(
	logger *slog.Logger,
	resource batch.ResourceWriter,
) *CsvWriter {
	return &CsvWriter{
		logger:   logger,
		resource: resource,
		encoded:  true,
	}
}

func (w *CsvWriter) WithFieldNames(fieldNames []string) *CsvWriter {
	w.fieldNames = fieldNames

	return w
}

func (w *CsvWriter) WithEncoded(encoded bool) *CsvWriter {
	w.encoded = encoded

	return w
}

var _ batch.Writer = (*CsvWriter)(nil)

func (w *CsvWriter) Open(ctx context.Context) error {
	if w.resource == nil {
		return &batch.IllegalArgumentError{Name: "resource"}
	}

	var errR error

	w.openOnce.Do(func() {
		err := w.resource.Open(ctx)
		if err != nil {
			errR = fmt.Errorf("could not open resource: %w", err)

			return
		}

		writer := csv.NewWriter(w.resource)

		if w.encoded {
			encoder := csvutil.NewEncoder(writer)

			if w.fieldNames != nil {
				encoder.SetHeader(w.fieldNames)
			}

			w.encoder = encoder
		}

		w.writer = writer
	})

	return errR
}

func (w *CsvWriter) Close(ctx context.Context) error {
	if w.resource == nil {
		w.logger.Warn("resource is not opened")

		return nil
	}

	var errR error

	w.closeOnce.Do(func() {
		if err := w.resource.Close(ctx); err != nil {
			errR = fmt.Errorf("could not close resource: %w", err)
		}

		w.encoder = nil

		w.writer = nil

		w.resource = nil
	})

	return errR
}

func (w *CsvWriter) Write(_ context.Context, record any) error {
	if w.encoder == nil {
		var typed []string

		switch typedRecord := record.(type) {
		case []string:
			typed = typedRecord
		case []interface{}:
			for _, item := range typedRecord {
				typedItem, cok := item.(string)
				if !cok {
					return &batch.IllegalArgumentError{Name: "record"}
				}

				typed = append(typed, typedItem)
			}
		default:
			return &batch.IllegalArgumentError{Name: "record"}
		}

		err := w.writer.Write(typed)
		if err != nil {
			return &batch.IoError{Operation: batch.IoWrite, Resource: w.resource.GetID(), Err: err}
		}
	} else {
		err := w.encoder.Encode(record)
		if err != nil {
			return &batch.IoError{Operation: batch.IoWrite, Resource: w.resource.GetID(), Err: err}
		}
	}

	w.writer.Flush()

	return nil
}
