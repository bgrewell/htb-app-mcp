// Package htb is the shared HTTP client for the HackTheBox main-app API.
//
// Domain-specific request/response types and methods live in sub-packages
// (internal/htb/machines, internal/htb/challenges, ...). Those sub-packages
// embed *Client and call its Do method.
//
// The client is deliberately small: it knows how to build authenticated
// requests, retry transient failures, honor rate-limit headers, and emit
// structured logs. It does not know about any HTB endpoint shapes.
package htb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the production HTB main-app API.
const DefaultBaseURL = "https://labs.hackthebox.com/api/v4"

// Config controls Client construction. All fields are optional except Token.
type Config struct {
	// Token is the HackTheBox App Token. Required. Never logged.
	Token string

	// BaseURL overrides DefaultBaseURL when non-empty.
	BaseURL string

	// HTTPClient is the underlying transport. Defaults to &http.Client{}
	// with Timeout=RequestTimeout.
	HTTPClient *http.Client

	// RequestTimeout is the per-request timeout. Defaults to 30s.
	RequestTimeout time.Duration

	// MaxRetries is the number of retry attempts after the first try.
	// Defaults to 3.
	MaxRetries int

	// Logger receives structured events. Defaults to slog.Default().
	Logger *slog.Logger

	// UserAgent overrides the default User-Agent header.
	UserAgent string
}

// Client is the shared, low-level HTTP client.
type Client struct {
	baseURL    *url.URL
	token      string
	http       *http.Client
	logger     *slog.Logger
	userAgent  string
	maxRetries int
}

// New returns a Client. It returns an error when Token is empty or BaseURL
// fails to parse.
func New(cfg Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, errors.New("htb: Config.Token is required")
	}

	rawURL := cfg.BaseURL
	if rawURL == "" {
		rawURL = DefaultBaseURL
	}
	base, err := url.Parse(strings.TrimRight(rawURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("htb: parsing BaseURL %q: %w", rawURL, err)
	}

	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	ua := cfg.UserAgent
	if ua == "" {
		ua = "htb-app-mcp/0.0.0-dev (+https://github.com/bgrewell/htb-app-mcp)"
	}

	retries := cfg.MaxRetries
	if retries < 0 {
		retries = 0
	} else if retries == 0 {
		retries = 3
	}

	return &Client{
		baseURL:    base,
		token:      cfg.Token,
		http:       httpClient,
		logger:     logger,
		userAgent:  ua,
		maxRetries: retries,
	}, nil
}

// BaseURL returns the configured base URL as a string.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// NewRequest builds an authenticated request. The path is joined onto
// BaseURL; a leading slash is allowed but optional.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	rel, err := url.Parse(strings.TrimLeft(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("htb: parsing path %q: %w", path, err)
	}
	full := c.baseURL.ResolveReference(rel)
	// ResolveReference drops the base path when rel is absolute-ish.
	// Force-join to avoid losing /api/v4 when callers pass "machine/list".
	if !strings.HasPrefix(full.Path, c.baseURL.Path) {
		full.Path = strings.TrimRight(c.baseURL.Path, "/") + "/" + strings.TrimLeft(rel.Path, "/")
	}

	req, err := http.NewRequestWithContext(ctx, method, full.String(), body)
	if err != nil {
		return nil, fmt.Errorf("htb: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do executes a request with retry-with-backoff on 429 and 5xx. The caller
// owns the response body and must close it on a non-nil response.
//
// Do never logs the Authorization header. Request/response logs include
// method, path, status, attempt, and elapsed time only.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		start := time.Now()
		resp, err := c.http.Do(req)
		elapsed := time.Since(start)

		if err != nil {
			lastErr = err
			c.logger.Warn("htb request error",
				"method", req.Method,
				"path", req.URL.Path,
				"attempt", attempt,
				"elapsed_ms", elapsed.Milliseconds(),
				"error", err,
			)
			if !shouldRetryErr(err) || attempt == c.maxRetries {
				return nil, fmt.Errorf("htb: %s %s: %w", req.Method, req.URL.Path, err)
			}
			sleepBeforeRetry(req.Context(), attempt, 0)
			continue
		}

		c.logger.Debug("htb response",
			"method", req.Method,
			"path", req.URL.Path,
			"status", resp.StatusCode,
			"attempt", attempt,
			"elapsed_ms", elapsed.Milliseconds(),
		)

		if !shouldRetryStatus(resp.StatusCode) || attempt == c.maxRetries {
			return resp, nil
		}

		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		// Drain and close so the connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		sleepBeforeRetry(req.Context(), attempt, retryAfter)
	}
	if lastErr != nil {
		return nil, fmt.Errorf("htb: %s %s: %w", req.Method, req.URL.Path, lastErr)
	}
	return nil, fmt.Errorf("htb: %s %s: exhausted %d retries", req.Method, req.URL.Path, c.maxRetries)
}

func shouldRetryStatus(code int) bool {
	if code == http.StatusTooManyRequests {
		return true
	}
	return code >= 500 && code <= 599
}

func shouldRetryErr(err error) bool {
	// Treat all network errors as retriable. Context cancellation is not.
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(strings.TrimSpace(h)); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

func sleepBeforeRetry(ctx context.Context, attempt int, hint time.Duration) {
	d := hint
	if d <= 0 {
		// Exponential backoff with jitter: 250ms, 500ms, 1s, 2s, ...
		base := time.Duration(250*(1<<attempt)) * time.Millisecond
		jitter := time.Duration(rand.Int63n(int64(base / 4)))
		d = base + jitter
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
	}
}
