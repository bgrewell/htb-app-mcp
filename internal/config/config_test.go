package config

import (
	"reflect"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("HTB_API_KEY", "")
	t.Setenv("HTB_API_BASE_URL", "")
	t.Setenv("HTB_LOG_LEVEL", "")
	t.Setenv("HTB_HTTP_TIMEOUT", "")

	c, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", c.LogLevel)
	}
	if c.RequestTimeout != 30*time.Second {
		t.Errorf("RequestTimeout = %v, want 30s", c.RequestTimeout)
	}
	if len(c.EnabledDomains) != 0 {
		t.Errorf("EnabledDomains = %v, want empty", c.EnabledDomains)
	}
}

func TestLoad_EnableValidatesAndSorts(t *testing.T) {
	c, err := Load("challenges, machines ,machines")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"challenges", "machines"}
	if !reflect.DeepEqual(c.EnabledDomains, want) {
		t.Errorf("EnabledDomains = %v, want %v", c.EnabledDomains, want)
	}
	if !c.IsEnabled("machines") || !c.IsEnabled("challenges") {
		t.Error("IsEnabled returned false for an enabled domain")
	}
	if c.IsEnabled("sherlocks") {
		t.Error("IsEnabled returned true for a disabled domain")
	}
}

func TestLoad_EnableRejectsUnknown(t *testing.T) {
	if _, err := Load("not-a-domain"); err == nil {
		t.Fatal("expected error for unknown domain")
	}
}

func TestLoad_RejectsBadTimeout(t *testing.T) {
	t.Setenv("HTB_HTTP_TIMEOUT", "abc")
	if _, err := Load(""); err == nil {
		t.Fatal("expected error for non-numeric HTB_HTTP_TIMEOUT")
	}
	t.Setenv("HTB_HTTP_TIMEOUT", "-1")
	if _, err := Load(""); err == nil {
		t.Fatal("expected error for non-positive HTB_HTTP_TIMEOUT")
	}
}

func TestLoad_RejectsBadLogLevel(t *testing.T) {
	t.Setenv("HTB_LOG_LEVEL", "loud")
	if _, err := Load(""); err == nil {
		t.Fatal("expected error for unknown HTB_LOG_LEVEL")
	}
}

func TestAllDomainsHasNoDuplicates(t *testing.T) {
	seen := map[string]struct{}{}
	for _, d := range AllDomains {
		if _, dup := seen[d]; dup {
			t.Errorf("duplicate domain %q in AllDomains", d)
		}
		seen[d] = struct{}{}
	}
}
