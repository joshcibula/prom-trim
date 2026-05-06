// Package prometheus provides a client for querying Prometheus HTTP API.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// QueryResult represents the response from a Prometheus instant query.
type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []Sample `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// Sample represents a single time series sample returned by Prometheus.
type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  [2]interface{}    `json:"value"`
}

// Client is a thin HTTP client for the Prometheus query API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Prometheus API client.
// baseURL should be the root URL of the Prometheus instance (e.g. "http://localhost:9090").
// timeout controls the per-request deadline.
func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("prometheus baseURL must not be empty")
	}
	// Validate the URL is parseable.
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid prometheus URL %q: %w", baseURL, err)
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// QueryInstant executes an instant PromQL query against the Prometheus API
// and returns the parsed result.
func (c *Client) QueryInstant(ctx context.Context, query string) (*QueryResult, error) {
	endpoint := c.baseURL + "/api/v1/query"

	params := url.Values{}
	params.Set("query", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, body)
	}

	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response JSON: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed (%s): %s", result.ErrorType, result.Error)
	}

	return &result, nil
}
