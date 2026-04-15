// Package fetch implements the "http_fetch" tool: a guarded HTTP client that
// retrieves a URL and returns its status, headers and (size-limited) body. By
// default it refuses to reach private or loopback addresses so that a model
// cannot use it to probe internal services.
package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fetcher performs guarded HTTP requests.
type Fetcher struct {
	client       *http.Client
	maxBytes     int64
	allowPrivate bool
}

// Option customises a Fetcher.
type Option func(*Fetcher)

// WithMaxBytes caps how much of a response body is read.
func WithMaxBytes(n int64) Option {
	return func(f *Fetcher) {
		if n > 0 {
			f.maxBytes = n
		}
	}
}

// WithTimeout sets the per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(f *Fetcher) {
		if d > 0 {
			f.client.Timeout = d
		}
	}
}

// WithAllowPrivate permits connections to private addresses. Use only in trusted
// environments; it disables the SSRF guard.
func WithAllowPrivate(allow bool) Option {
	return func(f *Fetcher) { f.allowPrivate = allow }
}

// NewFetcher builds a Fetcher with the given options.
func NewFetcher(opts ...Option) *Fetcher {
	f := &Fetcher{
		client:   &http.Client{Timeout: 30 * time.Second},
		maxBytes: 5 << 20, // 5 MiB
	}
	for _, opt := range opts {
		opt(f)
	}
	f.client.Transport = &http.Transport{
		DialContext:         safeDialer(f.allowPrivate),
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return f
}

// Result is the outcome of a fetch.
type Result struct {
	Status     int
	Header     http.Header
	Body       string
	Truncated  bool
	FinalURL   string
	DurationMS int64
}

// Do performs an HTTP request. Only http and https schemes are allowed.
func (f *Fetcher) Do(ctx context.Context, method, url string, headers map[string]string, body string) (*Result, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("only http and https URLs are supported")
	}
	if method == "" {
		method = http.MethodGet
	}
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mcpkit-fetch/1.0")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, f.maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	truncated := int64(len(data)) > f.maxBytes
	if truncated {
		data = data[:f.maxBytes]
	}
	return &Result{
		Status:     resp.StatusCode,
		Header:     resp.Header,
		Body:       string(data),
		Truncated:  truncated,
		FinalURL:   resp.Request.URL.String(),
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}
