package resource

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ivxivx/go-recon/batch"
)

const defaultReqTimeout = 10 * time.Second

type HTTPResource struct {
	logger         *slog.Logger
	httpClient     *http.Client
	url            string
	requestTimeout time.Duration

	openOnce sync.Once

	reader io.Reader
}

func NewHTTPResource(
	logger *slog.Logger,
	url string,
) *HTTPResource {
	return NewHTTPResourceWithHeaders(logger, url, nil)
}

func NewHTTPResourceWithHeaders(
	logger *slog.Logger,
	url string,
	headers map[string]string,
) *HTTPResource {
	var hdrs map[string]string
	if headers == nil {
		hdrs = make(map[string]string)
	} else {
		hdrs = headers
	}

	return &HTTPResource{
		logger: logger,
		httpClient: &http.Client{
			Transport: &HTTPInterceptor{
				Headers:  hdrs,
				Delegate: http.DefaultTransport,
			},
		},
		url:            url,
		requestTimeout: defaultReqTimeout,
	}
}

func (r *HTTPResource) WithRequestTimeout(timeout time.Duration) *HTTPResource {
	r.requestTimeout = timeout

	return r
}

var (
	_ batch.Resource = (*HTTPResource)(nil)
	_ io.Reader      = (*HTTPResource)(nil)
)

func (r *HTTPResource) GetID() string {
	return r.url
}

func (r *HTTPResource) Open(ctx context.Context) error {
	if r.reader != nil {
		r.logger.Warn("resource is already opened", slog.String("resource", r.url))

		return nil
	}

	var errR error

	r.openOnce.Do(func() {
		ctxt, cancel := context.WithTimeout(ctx, r.requestTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctxt, http.MethodGet, r.url, http.NoBody)
		if err != nil {
			errR = fmt.Errorf("could not create request for %s: %w", r.url, err)

			return
		}

		res, err := r.httpClient.Do(req)
		if err != nil {
			errR = &batch.ConnectionError{Operation: batch.ConnOpen, Address: r.url, Err: err}

			return
		}

		defer res.Body.Close()

		if res.StatusCode >= http.StatusMultipleChoices {
			errR = &batch.InvalidStatusError{StatusCode: res.StatusCode}

			return
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			errR = &batch.IoError{Operation: batch.IoRead, Resource: r.url, Err: err}

			return
		}

		reader := bytes.NewReader(data)

		r.reader = reader
	})

	return errR
}

func (r *HTTPResource) Close(_ context.Context) error {
	if r.reader == nil {
		r.logger.Info("resource is not opened, skip close", slog.String("resource", r.url))

		return nil
	}

	r.reader = nil

	return nil
}

func (r *HTTPResource) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.url, Err: ErrResourceNotOpened}
	}

	n, err = r.reader.Read(p)

	if err == nil || errors.Is(err, io.EOF) {
		return n, err
	}

	return 0, &batch.IoError{Operation: batch.IoRead, Resource: r.url, Err: err}
}
