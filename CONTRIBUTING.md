# Contributing to htb-app-mcp

Thanks for your interest. This project is being built in **phases** — each phase covers one HackTheBox domain end-to-end (capture → OpenAPI doc → MCP tools → user docs). If you'd like to help, find the [active milestone](https://github.com/bgrewell/htb-app-mcp/milestones) and pick an open issue tagged with that phase's label.

## Code of conduct

This project follows the [Contributor Covenant 2.1](./CODE_OF_CONDUCT.md). Be kind.

## Workflow

We use a strict **plan → issues → PR clusters → review → merge** loop:

1. **Plan.** Every phase has a written plan and a milestone.
2. **Issues.** Each endpoint or tool is its own issue, tagged with `type/endpoint-doc` or `type/mcp-tool` and the relevant `domain/*` label, assigned to the phase milestone.
3. **PR clusters.** PRs bundle a logical sub-group of issues from a single phase (e.g. "machines: lifecycle endpoints" closing 3–5 issues). Don't open a PR per issue — they're too small. Don't open a PR spanning multiple phases — they're too big.
4. **Review.** Every PR runs CI and gets a GitHub Copilot code review. Address review comments before requesting human review (or merging, when human review isn't required).
5. **Merge.** Squash-merge into `main`. Delete the branch.

## Local development

You need:

- Go 1.25+ ([install](https://go.dev/dl/)) — required by `modelcontextprotocol/go-sdk`
- `golangci-lint` ([install](https://golangci-lint.run/usage/install/))
- `gitleaks` ([install](https://github.com/gitleaks/gitleaks#installing)) — pre-commit secret scanning
- `redocly` CLI ([install](https://redocly.com/docs/cli/installation)) — OpenAPI linting
- A HackTheBox API token in `.env` (copy from `.env.example`)

For the API capture harness (Phase 1+):

- `mitmproxy` ([install](https://docs.mitmproxy.org/stable/overview-installation/))

Setup:

```sh
git clone git@github.com:bgrewell/htb-app-mcp.git
cd htb-app-mcp
cp .env.example .env
# edit .env, paste your HTB_API_KEY
go build ./...
go test ./...
```

## Secret hygiene

- **`.env` is gitignored.** Never commit it. `gitleaks` runs in CI to catch accidental commits.
- Captured API fixtures live under `scripts/capture/fixtures/` and are scrubbed of personal identifiers before commit. The raw mitmproxy output goes to `scripts/capture/raw/` which is also gitignored.
- If you suspect you've leaked a token, **rotate it immediately** at https://app.hackthebox.com/profile/settings and force-push history before anything else.

## Adding a new domain

The per-domain pattern is documented in [`docs/developers/adding-a-domain.md`](./docs/developers/adding-a-domain.md) (added in Phase 2). Briefly:

1. Capture the domain's endpoints with the capture harness.
2. Document them in `openapi/openapi.yaml` under the domain's tag.
3. Add typed client methods under `internal/htb/<domain>/`.
4. Add MCP tools under `internal/tools/<domain>/`.
5. Wire the domain into `internal/config` and `internal/server` so `--enable=<domain>` works.
6. Add `docs/users/<domain>.md` and `docs/developers/<domain>.md`.
7. Write unit tests using fixtures and (optionally) integration tests gated by `//go:build integration`.

## Code style

- `gofmt` and `goimports` enforced via `golangci-lint`.
- Comments are scarce. Code should explain itself; comments explain *why* when the *why* is non-obvious.
- No `interface{}` / `any` where a typed struct would do.
- Errors get context: `fmt.Errorf("listing machines: %w", err)`, never bare `return err` at API boundaries.

## Questions

Open a [discussion](https://github.com/bgrewell/htb-app-mcp/discussions) (once enabled) or an issue with the `type/question` label.
