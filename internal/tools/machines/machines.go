// Package machines registers the MCP tools for the machines domain on
// the shared MCP server. The tools wrap methods on
// internal/htb/machines, which is the typed HTTP client.
package machines

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bgrewell/htb-app-mcp/internal/config"
	"github.com/bgrewell/htb-app-mcp/internal/htb"
	htbmachines "github.com/bgrewell/htb-app-mcp/internal/htb/machines"
)

// Registrar is the server.DomainRegistrar for machines tools.
type Registrar struct{}

// Domain implements server.DomainRegistrar.
func (Registrar) Domain() string { return "machines" }

// Register implements server.DomainRegistrar. It installs the five
// list-and-info tools for the machines domain on the given MCP server.
func (Registrar) Register(s *mcp.Server, c *htb.Client, _ *config.Config) error {
	if s == nil {
		return fmt.Errorf("tools/machines: nil server")
	}
	if c == nil {
		return fmt.Errorf("tools/machines: nil HTB client")
	}
	mc := htbmachines.New(c)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_active_spawn",
		Description: "Get the caller's currently-spawned HackTheBox machine, or report that none is active. Returns id, name, IP, lab, VPN server, spawn timer.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, *htbmachines.SpawnInfo, error) {
		spawn, err := mc.GetActiveSpawn(ctx)
		if err != nil {
			return nil, nil, err
		}
		if spawn == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No machine is currently spawned."}},
			}, nil, nil
		}
		return nil, spawn, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_recommended",
		Description: "Get the caller's two recommended HackTheBox machine cards plus their categories (seasonal, staff_pick, recommended, ...).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, *htbmachines.Recommendations, error) {
		r, err := mc.GetRecommended(ctx)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_info",
		Description: "Get full info for a HackTheBox machine by its display name (case-sensitive, e.g. \"Cap\"). Returns the canonical numeric id alongside ~50 fields covering difficulty, owns, blood holders, the maker, and caller-relative state (isTodo, ownRank, authUserIn*Owns).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getInfoInput) (*mcp.CallToolResult, *htbmachines.MachineInfo, error) {
		info, err := mc.GetInfo(ctx, in.Name)
		return nil, info, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_walkthroughs",
		Description: "Get walkthroughs metadata for a HackTheBox machine by its numeric id (obtain the id from machines_get_info). Returns the official PDF metadata, official video metadata, and a list of community-authored writeup links.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in machineIDInput) (*mcp.CallToolResult, *htbmachines.WalkthroughsBundle, error) {
		w, err := mc.GetWalkthroughs(ctx, in.MachineID)
		return nil, w, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_list_reviews",
		Description: "List reviews for a HackTheBox machine by its numeric id, paginated (Laravel paginator). Defaults: per_page=15, sort_type=desc, sort_by=created_at.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listReviewsInput) (*mcp.CallToolResult, *htbmachines.ReviewPage, error) {
		opts := htbmachines.ListReviewsOptions{
			Page:     in.Page,
			PerPage:  in.PerPage,
			SortType: in.SortType,
			SortBy:   in.SortBy,
		}
		p, err := mc.ListReviews(ctx, in.MachineID, opts)
		return nil, p, err
	})

	return nil
}

// ---------- tool input shapes (jsonschema is auto-derived by the SDK) ----------

type noInput struct{}

type getInfoInput struct {
	Name string `json:"name" jsonschema:"the machine's display name (case-sensitive, e.g. Cap)"`
}

type machineIDInput struct {
	MachineID int64 `json:"machine_id" jsonschema:"numeric machine id, from machines_get_info"`
}

type listReviewsInput struct {
	MachineID int64    `json:"machine_id" jsonschema:"numeric machine id"`
	Page      int      `json:"page,omitempty" jsonschema:"1-based page number (default 1)"`
	PerPage   int      `json:"per_page,omitempty" jsonschema:"results per page (default 15)"`
	SortType  string   `json:"sort_type,omitempty" jsonschema:"sort direction: asc or desc (default desc)"`
	SortBy    []string `json:"sort_by,omitempty" jsonschema:"sort field names, repeatable; observed value: created_at"`
}
