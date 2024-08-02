package reader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/jszwec/csvutil"

	"github.com/ivxivx/go-recon/batch"
	rs "github.com/ivxivx/go-recon/batch/resource"
	tfer "github.com/ivxivx/go-recon/batch/transformer"
)

type CsvReader struct {
	logger       *slog.Logger
	resource     batch.ResourceReader
	trimSpace    bool
	transformers map[string]tfer.FieldTransformer
	skipNonExist bool

	openOnce  sync.Once
	closeOnce sync.Once

	decoder *csvutil.Decoder
}

func NewCsvReader(
	logger *slog.Logger,
	resource batch.ResourceReader,
) *CsvReader {
	return &CsvReader{
		logger:       logger,
		resource:     resource,
		trimSpace:    true,
		transformers: map[string]tfer.FieldTransformer{},
	}
}

func (r *CsvReader) WithTrimSpace(trim bool) *CsvReader {
	r.trimSpace = trim

	return r
}

func (r *CsvReader) WithTransformers(transformers map[string]tfer.FieldTransformer) *CsvReader {
	if transformers != nil {
		r.transformers = transformers
	}

	return r
}

func (r *CsvReader) WithSkipNonExist(skip bool) *CsvReader {
	r.skipNonExist = skip

	return r
}

var _ batch.Reader = (*CsvReader)(nil)

func (r *CsvReader) Open(ctx context.Context) error {
	if r.resource == nil {
		return &batch.IllegalArgumentError{Name: "resource"}
	}

	if r.decoder != nil {
		r.logger.Warn("resource is already opened", slog.String("resource", r.resource.GetID()))

		return nil
	}

	var errR error

	r.openOnce.Do(func() {
		err := r.resource.Open(ctx)
		if err != nil {
			var ioError *batch.IoError
			if errors.As(err, &ioError) && os.IsNotExist(unwrapRootCause(err)) && r.skipNonExist {
				r.logger.Info("resource does not exist, skip", slog.String("resource", r.resource.GetID()))

				return
			}

			errR = fmt.Errorf("could not open resource: %w", err)

			return
		}

		reader := csv.NewReader(r.resource)

		decoder, err := csvutil.NewDecoder(reader)
		if err != nil {
			errR = &BadFormatError{
				ResourceID: r.resource.GetID(),
				Err:        err,
			}

			return
		}

		decoder.Map = r.transformField

		r.decoder = decoder
	})

	return errR
}

func (r *CsvReader) transformField(field, col string, _ any) string {
	val := field

	if r.trimSpace {
		val = strings.TrimSpace(field)
	}

	transformer, found := r.transformers[col]
	if !found {
		return val
	}

	transformed, err := transformer.Transform(val)
	if err != nil {
		r.logger.Warn("could not transform field", slog.String("value", field), slog.String("name", col),
			slog.Any("error", err))

		return val
	}

	return transformed
}

func (r *CsvReader) Close(ctx context.Context) error {
	if r.decoder == nil {
		r.logger.Warn("resource is not opened")

		return nil
	}

	var errR error

	r.closeOnce.Do(func() {
		if err := r.resource.Close(ctx); err != nil {
			errR = fmt.Errorf("could not close resource: %w", err)
		}

		r.decoder = nil
	})

	return errR
}

func (r *CsvReader) Read(_ context.Context, record any) error {
	if r.decoder == nil {
		if r.skipNonExist {
			return io.EOF
		}

		return &batch.IoError{Operation: batch.IoRead, Resource: r.resource.GetID(), Err: rs.ErrResourceNotOpened}
	}

	err := r.decoder.Decode(record)

	if err == nil || errors.Is(err, io.EOF) {
		return err
	}

	return &batch.IoError{Operation: batch.IoRead, Resource: r.resource.GetID(), Err: err}
}
