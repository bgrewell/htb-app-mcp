// Command htb-app-mcp is an unofficial Model Context Protocol server for
// the HackTheBox main application.
//
// Usage:
//
//	htb-app-mcp --version
//	htb-app-mcp --list-domains
//	htb-app-mcp --enable=machines,challenges
//
// Configuration is read from the environment. See .env.example for the
// full list. The server speaks the MCP protocol over stdio and is
// intended to be launched by an MCP client (Claude Desktop, Claude Code,
// or any other client that can spawn a stdio server).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bgrewell/htb-app-mcp/internal/config"
	"github.com/bgrewell/htb-app-mcp/internal/htb"
	"github.com/bgrewell/htb-app-mcp/internal/server"
	machinestools "github.com/bgrewell/htb-app-mcp/internal/tools/machines"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "0.0.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "htb-app-mcp:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("htb-app-mcp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		showVersion = fs.Bool("version", false, "print version and exit")
		listDomains = fs.Bool("list-domains", false, "list known HackTheBox domains and exit")
		enable      = fs.String("enable", "", "comma-separated list of domains to enable (see --list-domains)")
	)

	fs.Usage = func() {
		out := fs.Output()
		_, _ = fmt.Fprintf(out, "Usage: %s [flags]\n\nFlags:\n", fs.Name())
		fs.PrintDefaults()
		_, _ = fmt.Fprintln(out, "\nEnvironment:")
		_, _ = fmt.Fprintln(out, "  HTB_API_KEY        Required for a real run. App Token from HTB profile.")
		_, _ = fmt.Fprintln(out, "  HTB_API_BASE_URL   Optional. Defaults to "+htb.DefaultBaseURL)
		_, _ = fmt.Fprintln(out, "  HTB_HTTP_TIMEOUT   Optional. Per-request timeout in seconds (default 30).")
		_, _ = fmt.Fprintln(out, "  HTB_LOG_LEVEL      Optional. debug|info|warn|error (default info).")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *showVersion {
		fmt.Println(version)
		return nil
	}
	if *listDomains {
		for _, d := range config.AllDomains {
			fmt.Println(d)
		}
		return nil
	}

	cfg, err := config.Load(*enable)
	if err != nil {
		return err
	}

	logger := newLogger(cfg.LogLevel)

	if cfg.APIKey == "" {
		return fmt.Errorf("HTB_API_KEY is not set; see %s --help", os.Args[0])
	}
	if len(cfg.EnabledDomains) == 0 {
		logger.Warn("no domains enabled (Phase 1 bootstrap); server will accept connections but expose zero tools",
			"hint", "pass --enable=<domain> (see --list-domains)")
	}

	httpClient, err := htb.New(htb.Config{
		Token:          cfg.APIKey,
		BaseURL:        cfg.BaseURL,
		RequestTimeout: cfg.RequestTimeout,
		Logger:         logger,
		UserAgent:      "htb-app-mcp/" + version + " (+https://github.com/bgrewell/htb-app-mcp)",
	})
	if err != nil {
		return err
	}

	srv, err := server.New(server.Options{
		Cfg:        cfg,
		HTTPClient: httpClient,
		Version:    version,
		Logger:     logger,
		Registrars: []server.DomainRegistrar{
			machinestools.Registrar{},
		},
	})
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return srv.Run(ctx)
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	// Logs go to stderr so they do not corrupt the MCP stdio stream.
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}
