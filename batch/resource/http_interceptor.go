package resource

import (
	"fmt"
	"net/http"
)

type HTTPInterceptor struct {
	Headers  map[string]string
	Delegate http.RoundTripper
}

func (ic *HTTPInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	ic.setHeaders(req)

	return ic.executeRequest(req)
}

func (ic *HTTPInterceptor) setHeaders(req *http.Request) {
	if ic.Headers != nil {
		for key, value := range ic.Headers {
			req.Header.Set(key, value)
		}
	}
}

func (ic *HTTPInterceptor) executeRequest(req *http.Request) (*http.Response, error) {
	response, err := ic.Delegate.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}
