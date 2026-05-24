// Package config loads runtime configuration from the environment and from
// the command-line --enable flag. It also owns the canonical list of
// HackTheBox domains the server knows about.
package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AllDomains is the canonical, ordered list of HackTheBox domains the
// project plans to expose. Entries are added here when a domain's
// scaffolding lands. A domain may appear in this list before it has any
// registered MCP tools.
var AllDomains = []string{
	"machines",
	"challenges",
	"sherlocks",
	"profile",
	"rankings",
	"tracks",
	"pro-labs",
	"fortresses",
	"seasons",
	"vpn",
	"search",
}

// Config is the resolved runtime configuration.
type Config struct {
	// APIKey is the HackTheBox App Token. Empty is valid only when the
	// caller passes --list-domains or --version; for a real run it is
	// required.
	APIKey string

	// BaseURL is the API base URL. Empty means "use the htb package default".
	BaseURL string

	// RequestTimeout is the per-request timeout for HTTP calls.
	RequestTimeout time.Duration

	// LogLevel is one of debug, info, warn, error.
	LogLevel string

	// EnabledDomains is the set of domains the user passed to --enable,
	// validated against AllDomains. Order is not significant.
	EnabledDomains []string
}

// Load builds a Config from the environment and from a comma-separated
// --enable string. Validation rules:
//   - enableCSV may be empty (zero enabled domains, useful for --list-domains).
//   - Any unknown domain name in enableCSV is a hard error.
//   - HTB_API_KEY is NOT required by Load itself; callers check it after
//     deciding whether the run actually needs the API.
func Load(enableCSV string) (*Config, error) {
	c := &Config{
		APIKey:   os.Getenv("HTB_API_KEY"),
		BaseURL:  os.Getenv("HTB_API_BASE_URL"),
		LogLevel: getenvDefault("HTB_LOG_LEVEL", "info"),
	}

	timeoutStr := os.Getenv("HTB_HTTP_TIMEOUT")
	if timeoutStr == "" {
		c.RequestTimeout = 30 * time.Second
	} else {
		secs, err := strconv.Atoi(timeoutStr)
		if err != nil || secs <= 0 {
			return nil, fmt.Errorf("config: HTB_HTTP_TIMEOUT must be a positive integer (seconds), got %q", timeoutStr)
		}
		c.RequestTimeout = time.Duration(secs) * time.Second
	}

	domains, err := parseEnableCSV(enableCSV)
	if err != nil {
		return nil, err
	}
	c.EnabledDomains = domains

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, fmt.Errorf("config: HTB_LOG_LEVEL must be one of debug|info|warn|error, got %q", c.LogLevel)
	}

	return c, nil
}

// IsEnabled reports whether the given domain is in EnabledDomains.
func (c *Config) IsEnabled(domain string) bool {
	for _, d := range c.EnabledDomains {
		if d == domain {
			return true
		}
	}
	return false
}

func parseEnableCSV(s string) ([]string, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	known := make(map[string]struct{}, len(AllDomains))
	for _, d := range AllDomains {
		known[d] = struct{}{}
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, raw := range strings.Split(s, ",") {
		d := strings.TrimSpace(raw)
		if d == "" {
			continue
		}
		if _, ok := known[d]; !ok {
			return nil, fmt.Errorf("config: unknown domain %q in --enable; known domains: %s",
				d, strings.Join(AllDomains, ", "))
		}
		if _, dup := seen[d]; dup {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	sort.Strings(out)
	return out, nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
