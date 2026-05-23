# Copilot review instructions

GitHub Copilot reviews every PR in this repo before human review. Use these conventions when reviewing.

## Project shape

- Go MCP server built on `github.com/modelcontextprotocol/go-sdk`.
- Reverse-engineered OpenAPI 3.1 spec at `openapi/openapi.yaml` is canonical. If client code disagrees with the spec, the spec wins.
- Work is phased by domain. See `README.md` for the phase table. Reject changes that introduce code for a domain outside the active phase.

## What to look for

- **Secret handling.** `HTB_API_KEY` must come from env only. It must never appear in logs, error messages, MCP tool responses, or committed fixtures. Flag any path where a token could leak.
- **Fixture hygiene.** Captured fixtures under `scripts/capture/fixtures/` must have personal identifiers (usernames, emails, tokens, internal IDs) scrubbed. Raw captures belong in `scripts/capture/raw/` (gitignored).
- **OpenAPI parity.** Every new client method or MCP tool should match an `operationId` in `openapi/openapi.yaml`. Flag drift between the two.
- **Tool-count discipline.** Prefer one tool with a `kind` parameter over multiple near-duplicate tools. The active tool count lives in `docs/developers/tool-count.md` — flag PRs that change the count without updating that file.
- **Error wrapping.** API-boundary errors should be wrapped with context (`fmt.Errorf("listing machines: %w", err)`), never returned bare.
- **No speculative scope.** Reject abstractions, helpers, or feature flags introduced for hypothetical future use.

## Style

- `gofmt` and `goimports` are enforced via `golangci-lint`. CI catches formatting; skip style nits.
- Comments should explain *why*, not *what*. Flag comments that just restate the code.
- Prefer concrete types over `interface{}` / `any`.

## PR shape

- PRs should cluster a logical sub-group of issues from one phase. Single-issue PRs and multi-phase PRs are both anti-patterns here — call them out.
- Every endpoint or tool change should reference the closing issue(s) in the PR body.
