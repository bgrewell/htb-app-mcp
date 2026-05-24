# htb-app-mcp

> **Status: pre-alpha.** Active scaffolding. Not yet usable.

An unofficial **Model Context Protocol** server for the [HackTheBox](https://www.hackthebox.com) main application — machines, challenges, sherlocks, profiles, rankings, tracks, pro labs, fortresses, seasons, VPN, search.

HackTheBox ships an official MCP server for their [CTF platform](https://help.hackthebox.com/en/articles/11793915-model-context-protocol-for-ctf), but there is no official MCP or public API documentation for the main app. This project fills both gaps:

1. A reverse-engineered **[OpenAPI 3.1 spec](./openapi/openapi.yaml)** for the HTB main-app HTTP API, rendered as a Swagger UI site (see [docs site](https://bgrewell.github.io/htb-app-mcp/) — published once Phase 1 lands).
2. A **Go MCP server** that exposes the documented endpoints as MCP tools, usable from Claude Desktop, Claude Code, and any other MCP client.

## Goals

- 100% coverage of the HackTheBox main app surface, delivered in domain-scoped phases (machines → challenges → sherlocks → everything else).
- One binary, with **per-domain enable flags** so users only load the tool surface they actually need (`--enable=machines,challenges`). Keeps active tool counts low without forcing multiple installs.
- High-quality docs for three audiences — end users, contributors, and LLMs (see [`docs/`](./docs/)).
- Canonical OpenAPI spec as a standalone deliverable, useful to anyone building HTB tooling.

## Project status

This repo is in **Phase 0 (foundation)**. See [the roadmap](#roadmap) and the [open milestones](https://github.com/bgrewell/htb-app-mcp/milestones) for what's planned and what's done.

| Phase | Scope | Status |
|------|------|------|
| 0 — Foundation | repo, CI, docs scaffolding | in progress |
| 1 — API recon + OpenAPI foundation | capture harness, client, MCP bootstrap | not started |
| 2 — Machines domain | list, info, lifecycle, flags | not started |
| 3 — Challenges domain | | not started |
| 4 — Sherlocks domain | | not started |
| 5 — Profile / Rankings / Tracks / Pro Labs / Fortresses / Seasons / VPN / Search | | not started |
| 6 — v1.0 release + distribution | GoReleaser, Docker, MCP registry | not started |

## Quick start

> Not yet usable. This section will be filled in as Phase 2 lands.

```sh
# Pre-alpha; APIs and flags will change.
go install github.com/bgrewell/htb-app-mcp/cmd/htb-app-mcp@latest

export HTB_API_KEY=<your token from app.hackthebox.com profile settings>
htb-app-mcp --enable=machines
```

## Documentation

- **[User docs](./docs/users/)** — install, configuration, per-domain tool reference.
- **[Developer docs](./docs/developers/)** — architecture, contribution workflow, adding a new domain.
- **[LLM docs](./docs/llms/)** and **[CLAUDE.md](./CLAUDE.md)** — conventions for AI assistants working in this repo.
- **[API reverse-engineering notes](./docs/api/)** — capture workflow, endpoint inventory.
- **[OpenAPI spec](./openapi/openapi.yaml)** — canonical API documentation.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Work is tracked via issues; PRs cluster related issues by phase.

## Prior art & credits

- HackTheBox for the platform and for offering user-generated API tokens that make community tooling possible.
- [`noaslr/htb-mcp-server`](https://github.com/noaslr/htb-mcp-server) — a Go MCP for a subset of the main app, used as structural reference.
- [`Propolisa/htb-api-docs`](https://github.com/Propolisa/htb-api-docs) — community-maintained Postman collection, useful as a hint of which endpoints exist. We do not import schemas from it; every operation in our OpenAPI spec is verified by a fresh capture against the live API.
- [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) — the official Go MCP SDK this server is built on.

This project is **not affiliated with or endorsed by HackTheBox**.

## License

MIT — see [LICENSE](./LICENSE).
