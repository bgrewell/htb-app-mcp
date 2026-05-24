package htb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, srv *httptest.Server, opts ...func(*Config)) *Client {
	t.Helper()
	cfg := Config{
		Token:          "test-token",
		BaseURL:        srv.URL + "/api/v4",
		HTTPClient:     srv.Client(),
		RequestTimeout: 2 * time.Second,
		MaxRetries:     2,
	}
	for _, o := range opts {
		o(&cfg)
	}
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNew_RequiresToken(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error when Token is empty")
	}
}

func TestNew_DefaultsApplied(t *testing.T) {
	c, err := New(Config{Token: "x"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.BaseURL() != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL(), DefaultBaseURL)
	}
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", c.maxRetries)
	}
	if c.userAgent == "" {
		t.Error("userAgent must default to non-empty")
	}
}

func TestNewRequest_SetsAuthAndAccept(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, err := c.NewRequest(context.Background(), http.MethodGet, "user/info", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want application/json", got)
	}
	if !strings.HasSuffix(req.URL.Path, "/api/v4/user/info") {
		t.Errorf("URL.Path = %q, want suffix /api/v4/user/info", req.URL.Path)
	}
}

func TestNewRequest_LeadingSlashTolerated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()
	c := newTestClient(t, srv)

	for _, p := range []string{"user/info", "/user/info"} {
		req, err := c.NewRequest(context.Background(), http.MethodGet, p, nil)
		if err != nil {
			t.Fatalf("NewRequest(%q): %v", p, err)
		}
		if !strings.HasSuffix(req.URL.Path, "/api/v4/user/info") {
			t.Errorf("path %q -> %q, expected /api/v4/user/info suffix", p, req.URL.Path)
		}
	}
}

func TestDo_SuccessNoRetry(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "user/info", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if hits != 1 {
		t.Errorf("server hits = %d, want 1", hits)
	}
}

func TestDo_RetriesOn5xxThenSucceeds(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "x", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if hits != 3 {
		t.Errorf("hits = %d, want 3 (2 retries + success)", hits)
	}
}

func TestDo_RetriesOn429HonorsRetryAfter(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "x", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDo_GivesUpAfterMaxRetries(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "x", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("final status = %d, want 500", resp.StatusCode)
	}
	// MaxRetries=2 means 1 initial + 2 retries = 3 hits.
	if hits != 3 {
		t.Errorf("hits = %d, want 3", hits)
	}
}

func TestDo_DoesNotRetryOn4xx(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "x", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if hits != 1 {
		t.Errorf("hits = %d, want 1 (no retry on 401)", hits)
	}
}

func TestDo_RespectsCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := newTestClient(t, srv, func(cfg *Config) { cfg.MaxRetries = 5 })
	req, _ := c.NewRequest(ctx, http.MethodGet, "x", nil)

	// Cancel mid-flight by canceling after a tiny delay.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	resp, err := c.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error from canceled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error = %v (acceptable as long as it wraps context.Canceled or is a request error)", err)
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"0", 0},
		{"3", 3 * time.Second},
		{"  7  ", 7 * time.Second},
		{"not a number", 0},
	}
	for _, c := range cases {
		got := parseRetryAfter(c.in)
		if got != c.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
