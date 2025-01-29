package http

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
)

// Config holds the HTTP client configuration
type Config struct {
	// Timeout is the maximum time to wait for a request to complete
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// MaxRetries is the maximum number of retries for a request
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	// RetryWaitMin is the minimum time to wait between retries
	RetryWaitMin time.Duration `json:"retry_wait_min" yaml:"retry_wait_min"`
	// RetryWaitMax is the maximum time to wait between retries
	RetryWaitMax time.Duration `json:"retry_wait_max" yaml:"retry_wait_max"`
	// RetryableStatusCodes are HTTP status codes that should trigger a retry
	RetryableStatusCodes []int `json:"retryable_status_codes" yaml:"retryable_status_codes"`
}

// DefaultConfig returns the default HTTP client configuration
func DefaultConfig() Config {
	return Config{
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		RetryWaitMin:         1 * time.Second,
		RetryWaitMax:         5 * time.Second,
		RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
	}
}

// Client is an HTTP client with retry and metrics capabilities
type Client struct {
	client  *http.Client
	config  Config
	metrics *MetricsReporter
}

// New creates a new HTTP client
func New(config Config, metricsReporter *metrics.Reporter) *Client {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &Client{
		client:  client,
		config:  config,
		metrics: NewMetricsReporter(metricsReporter),
	}
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	var resp *http.Response
	var err error

	start := time.Now()
	defer func() {
		c.metrics.ObserveRequest(req.Method, resp, err, time.Since(start))
	}()

	for i := 0; i <= c.config.MaxRetries; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			c.metrics.ObserveError("request_failed")
			if i == c.config.MaxRetries {
				return nil, fmt.Errorf("max retries reached: %w", err)
			}
			continue
		}

		if !c.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		if i == c.config.MaxRetries {
			return resp, nil
		}

		// Close the response body before retrying
		if resp.Body != nil {
			resp.Body.Close()
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.getRetryBackoff(i)):
			continue
		}
	}

	return resp, err
}

// Get sends a GET request to the specified URL
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post sends a POST request to the specified URL
func (c *Client) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// shouldRetry returns true if the status code is retryable
func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.config.RetryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// getRetryBackoff returns the backoff duration for a retry attempt
func (c *Client) getRetryBackoff(attempt int) time.Duration {
	// Simple exponential backoff with jitter
	backoff := c.config.RetryWaitMin.Seconds() * math.Pow(2, float64(attempt))
	if backoff > c.config.RetryWaitMax.Seconds() {
		backoff = c.config.RetryWaitMax.Seconds()
	}
	// Add jitter by varying the backoff by ±25%
	jitter := (backoff * 0.5) - (backoff * 0.25)
	return time.Duration((backoff + jitter) * float64(time.Second))
}
