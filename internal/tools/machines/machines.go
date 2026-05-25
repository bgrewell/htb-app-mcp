// Package machines registers the MCP tools for the machines domain on
// the shared MCP server. The tools wrap methods on
// internal/htb/machines, which is the typed HTTP client.
package machines

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_list_walkthrough_languages",
		Description: "List the language enum used to tag community walkthroughs. Useful when the caller wants to filter walkthroughs by language.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, *languagesOut, error) {
		ls, err := mc.ListWalkthroughLanguages(ctx)
		if err != nil {
			return nil, nil, err
		}
		return nil, &languagesOut{Languages: ls}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_random_walkthrough_machine",
		Description: "Get a random HackTheBox machine that has community walkthroughs. Returns a minimal reference (id, name, avatar). Caller typically follows up with machines_get_walkthroughs.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, *htbmachines.MachineRef, error) {
		r, err := mc.GetRandomWalkthroughMachine(ctx)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_get_graph_matrix",
		Description: "Get the difficulty radar matrix for a HackTheBox machine by its numeric id. Returns scores across five skill axes (ctf, custom, cve, enum, real) for aggregate, maker, and the caller themselves.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in machineIDInput) (*mcp.CallToolResult, *htbmachines.GraphMatrix, error) {
		gm, err := mc.GetGraphMatrix(ctx, in.MachineID)
		return nil, gm, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_list_tasks",
		Description: "List guided-mode tasks for a HackTheBox machine by its numeric id. SENSITIVITY: the `flag` field contains the actual plaintext flag value for any task the caller has already completed. Tasks chain via `prerequisite_id`.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in machineIDInput) (*mcp.CallToolResult, *tasksOut, error) {
		ts, err := mc.ListTasks(ctx, in.MachineID)
		if err != nil {
			return nil, nil, err
		}
		return nil, &tasksOut{Tasks: ts}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_list_adventure_steps",
		Description: "List adventure-mode steps for a HackTheBox machine by its numeric id. Adventure compresses the machine to canonical flag-submission steps (typically Submit User Flag, Submit Root Flag).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in machineIDInput) (*mcp.CallToolResult, *adventureOut, error) {
		ss, err := mc.ListAdventureSteps(ctx, in.MachineID)
		if err != nil {
			return nil, nil, err
		}
		return nil, &adventureOut{Steps: ss}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "machines_save_official_writeup_pdf",
		Description: "Download the official PDF writeup for a HackTheBox machine and save it to a directory on the local filesystem. The directory must exist and be writable. Returns the full path of the saved file. Useful for bulk-downloading writeups to a local library.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in saveWriteupInput) (*mcp.CallToolResult, *saveWriteupOut, error) {
		return saveWriteupHandler(ctx, mc, in)
	})

	return nil
}

// saveWriteupHandler validates the output directory, opens a destination
// file, streams the PDF body into it, and returns the saved path.
func saveWriteupHandler(ctx context.Context, mc *htbmachines.Client, in saveWriteupInput) (*mcp.CallToolResult, *saveWriteupOut, error) {
	if in.OutputDir == "" {
		return nil, nil, fmt.Errorf("output_dir is required")
	}
	if in.MachineID <= 0 {
		return nil, nil, fmt.Errorf("machine_id must be positive")
	}
	info, err := os.Stat(in.OutputDir)
	if err != nil {
		return nil, nil, fmt.Errorf("output_dir %q: %w", in.OutputDir, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("output_dir %q is not a directory", in.OutputDir)
	}
	name := in.Filename
	if name == "" {
		name = "machine_" + strconv.FormatInt(in.MachineID, 10) + ".pdf"
	}
	out := filepath.Join(in.OutputDir, name)
	f, err := os.Create(out)
	if err != nil {
		return nil, nil, fmt.Errorf("create %q: %w", out, err)
	}
	n, copyErr := mc.DownloadOfficialWriteupPDF(ctx, in.MachineID, f)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(out)
		return nil, nil, copyErr
	}
	if closeErr != nil {
		return nil, nil, fmt.Errorf("close %q: %w", out, closeErr)
	}
	return nil, &saveWriteupOut{Path: out, Bytes: n}, nil
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

type saveWriteupInput struct {
	MachineID int64  `json:"machine_id" jsonschema:"numeric machine id"`
	OutputDir string `json:"output_dir" jsonschema:"absolute path to an existing writable directory where the PDF will be saved"`
	Filename  string `json:"filename,omitempty" jsonschema:"optional filename for the PDF; defaults to machine_<id>.pdf"`
}

// Output wrappers — the SDK requires a struct (not a bare slice) for
// structured tool output so the JSON Schema has a named root.
type languagesOut struct {
	Languages []htbmachines.Language `json:"languages"`
}
type tasksOut struct {
	Tasks []htbmachines.MachineTask `json:"tasks"`
}
type adventureOut struct {
	Steps []htbmachines.AdventureStep `json:"steps"`
}
type saveWriteupOut struct {
	Path  string `json:"path" jsonschema:"absolute path of the saved PDF"`
	Bytes int64  `json:"bytes" jsonschema:"number of bytes written"`
}
