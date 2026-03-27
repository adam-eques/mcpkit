// Package websearch implements the "web_search" tool backed by the DuckDuckGo
// Instant Answer API. The endpoint is injectable so the tool can be exercised in
// tests without reaching the network.
package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DefaultEndpoint is the public DuckDuckGo Instant Answer API.
const DefaultEndpoint = "https://api.duckduckgo.com/"

// Searcher queries an Instant Answer endpoint.
type Searcher struct {
	endpoint string
	client   *http.Client
}

// Option customises a Searcher.
type Option func(*Searcher)

// WithEndpoint overrides the API endpoint, primarily for testing.
func WithEndpoint(u string) Option { return func(s *Searcher) { s.endpoint = u } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(s *Searcher) { s.client = c } }

// NewSearcher builds a Searcher.
func NewSearcher(opts ...Option) *Searcher {
	s := &Searcher{
		endpoint: DefaultEndpoint,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Result is a single search result.
type Result struct {
	Title string
	URL   string
}

// apiResponse mirrors the subset of the Instant Answer schema we consume.
type apiResponse struct {
	AbstractText  string `json:"AbstractText"`
	AbstractURL   string `json:"AbstractURL"`
	Heading       string `json:"Heading"`
	RelatedTopics []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}

// Search returns the abstract (if any) followed by related topics.
func (s *Searcher) Search(ctx context.Context, query string, limit int) (string, []Result, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("no_html", "1")
	q.Set("skip_disambig", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", "mcpkit-websearch/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("search endpoint returned %d", resp.StatusCode)
	}

	var body apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", nil, err
	}

	results := make([]Result, 0, limit)
	for _, rt := range body.RelatedTopics {
		if rt.Text == "" || rt.FirstURL == "" {
			continue
		}
		results = append(results, Result{Title: rt.Text, URL: rt.FirstURL})
		if len(results) >= limit {
			break
		}
	}
	return body.AbstractText, results, nil
}
