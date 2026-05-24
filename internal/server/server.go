// Package server wires the MCP server, the shared HTB HTTP client, and
// the per-domain tool registrars together. Domain tool packages register
// themselves by calling Register here from cmd/htb-app-mcp/main.go.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bgrewell/htb-app-mcp/internal/config"
	"github.com/bgrewell/htb-app-mcp/internal/htb"
)

// Name and Version are wired into the MCP Implementation block returned
// to clients during the initialize handshake.
const Name = "htb-app-mcp"

// DomainRegistrar registers all MCP tools for one HackTheBox domain on
// the given server. Implementations live in internal/tools/<domain>/.
type DomainRegistrar interface {
	// Domain returns the canonical short name, matching one of config.AllDomains.
	Domain() string
	// Register installs the domain's tools on the server.
	Register(s *mcp.Server, c *htb.Client, cfg *config.Config) error
}

// Options configures a new server. Logger is required; Version defaults
// to "0.0.0-dev".
type Options struct {
	Cfg        *config.Config
	HTTPClient *htb.Client
	Version    string
	Logger     *slog.Logger
	// Registrars is the full set of domain registrars compiled into the
	// binary. Server.Register only calls Register on the ones whose
	// Domain() is in Cfg.EnabledDomains.
	Registrars []DomainRegistrar
}

// Server is the runnable MCP server.
type Server struct {
	mcp     *mcp.Server
	cfg     *config.Config
	htb     *htb.Client
	logger  *slog.Logger
	regs    []DomainRegistrar
	version string
}

// New builds a Server, validates that every enabled domain has a
// registrar, and installs the tools for those domains. Returns an error
// if an enabled domain has no registrar OR if any registrar fails to
// install.
func New(opts Options) (*Server, error) {
	if opts.Cfg == nil {
		return nil, errors.New("server: Options.Cfg is required")
	}
	if opts.Logger == nil {
		return nil, errors.New("server: Options.Logger is required")
	}
	version := opts.Version
	if version == "" {
		version = "0.0.0-dev"
	}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    Name,
		Version: version,
	}, nil)

	s := &Server{
		mcp:     mcpServer,
		cfg:     opts.Cfg,
		htb:     opts.HTTPClient,
		logger:  opts.Logger,
		regs:    opts.Registrars,
		version: version,
	}

	byDomain := make(map[string]DomainRegistrar, len(opts.Registrars))
	for _, r := range opts.Registrars {
		byDomain[r.Domain()] = r
	}

	for _, d := range opts.Cfg.EnabledDomains {
		r, ok := byDomain[d]
		if !ok {
			s.logger.Warn("enabled domain has no registered tools (Phase 1 bootstrap)",
				"domain", d)
			continue
		}
		if opts.HTTPClient == nil {
			return nil, fmt.Errorf("server: cannot register domain %q without an HTB client", d)
		}
		if err := r.Register(mcpServer, opts.HTTPClient, opts.Cfg); err != nil {
			return nil, fmt.Errorf("server: registering domain %q: %w", d, err)
		}
		s.logger.Info("registered domain tools", "domain", d)
	}

	return s, nil
}

// Run serves the MCP protocol over stdio until ctx is canceled or the
// peer disconnects.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("htb-app-mcp starting",
		"version", s.version,
		"enabled_domains", s.cfg.EnabledDomains,
	)
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}
