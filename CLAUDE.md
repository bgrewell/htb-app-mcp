# CLAUDE.md

Single source of context for Claude and other LLM coding assistants working in this repo. Keep this file authoritative — if it disagrees with the code, update the file.

## What this project is

An unofficial Model Context Protocol server for the HackTheBox main application, plus a reverse-engineered OpenAPI 3.1 spec for the same API. The spec is a deliverable in its own right.

See `README.md` for user-facing summary and the phase roadmap.

## Repository layout

```
cmd/htb-app-mcp/        # binary entrypoint
internal/config/        # env loading, --enable flag parsing
internal/htb/           # HTTP client, one sub-package per domain
internal/server/        # MCP server bootstrap, transport selection
internal/tools/         # MCP tool registration, one sub-package per domain
openapi/openapi.yaml    # canonical OpenAPI 3.1 spec
openapi/components/     # split schemas when openapi.yaml grows
docs/users/             # end-user install + per-domain tool reference
docs/developers/        # contributor + architecture docs
docs/llms/              # LLM-focused conventions (supplements this file)
docs/api/               # reverse-engineering notes per domain
site/                   # Swagger UI assets
scripts/capture/        # mitmproxy harness + recorded fixtures
.github/workflows/      # CI, Pages, release
```

Anything outside `cmd/` and `internal/` is documentation or tooling.

## Phased delivery

Work is delivered in phases, one HackTheBox domain at a time:

| Phase | Scope |
|-------|-------|
| 0 | Foundation — repo, CI, scaffolding |
| 1 | API recon, OpenAPI scaffolding, capture harness, MCP bootstrap |
| 2 | Machines domain (template for 3–5) |
| 3 | Challenges |
| 4 | Sherlocks |
| 5 | Profile, Rankings, Tracks, Pro Labs, Fortresses, Seasons, VPN, Search |
| 6 | v1.0 release + distribution |

**Do not write code for a domain that is not in the active phase.** Issues and milestones, yes; code, no.

## Conventions

### Go

- Go 1.25+ (required by `modelcontextprotocol/go-sdk`).
- `gofmt` and `goimports` enforced via `golangci-lint`.
- No `interface{}` / `any` where a typed struct would do.
- Errors wrap with context: `fmt.Errorf("listing machines: %w", err)`. Never `return err` at API boundaries.
- Comments are scarce. The code should explain itself; comments explain *why* when the *why* is non-obvious.
- One file per logical sub-group inside a domain package (e.g. `internal/htb/machines/list.go`, `lifecycle.go`, `flags.go`).

### OpenAPI

- `openapi/openapi.yaml` is canonical. If the spec and the client disagree, the spec wins and the client is fixed.
- **Source of truth is our own captures, full stop.** Third-party Postman collections, community docs, and prior MCP repos are hints — they tell you which endpoints exist, not what their shape is. Every operation we document must be backed by a fresh capture under `scripts/capture/fixtures/<domain>/`. Do not copy schemas from external sources.
- Each endpoint gets at least one re-capture at a later date to confirm shape stability.
- One tag per domain. Every operation has `operationId` in `camelCase` matching the Go client method name.
- Every operation has at least one example response, sourced from a captured fixture.
- Run `redocly lint openapi/openapi.yaml` before committing changes to the spec.
- See `docs/developers/openapi-conventions.md` for the full convention set.

### MCP tools

- Every tool added must justify itself. Prefer one tool that takes a `kind` parameter over three near-duplicates.
- Tools are registered via a `Register(server, client, config)` function in each `internal/tools/<domain>/` package, called from `internal/server` only when `<domain>` is in the `--enable` list.
- Tool names use `snake_case` and start with the domain: `machines_list`, `machines_spawn`, `challenges_info`.
- Track active tool count in `docs/developers/tool-count.md`. The MCP accuracy ceiling sits around 25 tools loaded at once — design with that in mind.

### Tests

- Unit tests use `httptest` + recorded fixtures under `scripts/capture/fixtures/<domain>/`.
- Integration tests that hit the real API are gated by `//go:build integration` and require `HTB_API_KEY` in the environment.

### Secret hygiene

- `.env` is `.gitignored`. The `HTB_API_KEY` it contains must never be staged.
- `gitleaks` runs in CI to catch accidental commits.
- Captured fixtures are scrubbed of personal identifiers before commit. Raw mitmproxy output goes to `scripts/capture/raw/` (gitignored).
- Tokens are read from env only. Never log them. Never echo them in tool responses.

## Adding a new domain (Phase 2 onward)

The full pattern is documented in `docs/developers/adding-a-domain.md` once Phase 2 lands. Outline:

1. Capture the domain's endpoints with the harness in `scripts/capture/`.
2. Document them in `openapi/openapi.yaml` under the domain's tag, with examples drawn from fixtures.
3. Add typed client methods under `internal/htb/<domain>/`.
4. Add MCP tools under `internal/tools/<domain>/`.
5. Wire the domain into `internal/config` and `internal/server` so `--enable=<domain>` works.
6. Add `docs/users/<domain>.md` and `docs/developers/<domain>.md`.
7. Write unit tests against fixtures; optionally add integration tests behind the build tag.

## Workflow

- Each endpoint or tool is one GitHub issue tagged `type/endpoint-doc` or `type/mcp-tool` plus the relevant `domain/*` label, assigned to the phase milestone.
- PRs bundle a logical sub-group of issues from a single phase. Not one PR per issue, not one PR spanning multiple phases.
- Every PR runs CI and gets a GitHub Copilot review. Address Copilot comments before requesting human review or merging.
- Squash-merge into `main`. Delete the branch.

## What not to do

- Don't add features, abstractions, or scope beyond what the current task requires.
- Don't write code for domains not in the active phase.
- Don't commit `.env`, raw captures, or anything containing a token or username.
- Don't add `Co-Authored-By` trailers or "Generated with Claude Code" PR footers — the user has explicitly disabled these.
- Don't use `--no-verify`, `--no-gpg-sign`, or otherwise bypass commit hooks.
- Don't open a PR per issue. Cluster issues into a single phase-scoped PR.
