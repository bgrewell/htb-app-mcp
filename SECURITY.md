# Security policy

## Reporting a vulnerability

If you believe you've found a vulnerability in this project, please report it privately rather than opening a public issue.

- Open a [GitHub private vulnerability report](https://github.com/bgrewell/htb-app-mcp/security/advisories/new), **or**
- Email the maintainer: `bgrewell@gmail.com` with subject prefix `[htb-app-mcp security]`.

Please include:

- A description of the issue and where it lives in the code.
- Steps to reproduce, if possible.
- The impact you believe it has.

You can expect an initial acknowledgement within 7 days. Fixes will be coordinated privately and disclosed after a patched release is available.

## Scope

This project is a client for the HackTheBox main-app HTTP API. It is **not affiliated with or endorsed by HackTheBox**.

In scope:

- Bugs in this codebase that could leak the user's `HTB_API_KEY` (logs, error messages, tool responses, on-disk fixtures, etc.).
- Bugs that cause the MCP server to perform actions the user did not authorize via a tool call.
- Bugs in the capture harness (`scripts/capture/`) that could commit personal identifiers or tokens to the repo.
- Dependency vulnerabilities surfaced by `go mod` / GitHub's Dependabot.

Out of scope:

- Vulnerabilities in the HackTheBox platform itself — report those to HackTheBox via [their disclosure process](https://www.hackthebox.com/contact).
- Misuse of valid API tokens by their owner.
- Rate-limit handling on the HTB API.

## Token handling

This server requires a HackTheBox API token (`HTB_API_KEY`). Treat it like a password:

- Tokens are read from environment variables only — never from command-line flags.
- Tokens are never written to logs, even at debug level.
- Tokens are never echoed back through MCP tool responses.
- `.env` is `.gitignored` from commit #1 and `gitleaks` runs in CI.

If you suspect your token has been exposed, rotate it immediately at https://app.hackthebox.com/profile/settings.
