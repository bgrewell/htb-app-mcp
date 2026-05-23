# API capture workflow

The HackTheBox main-app API is undocumented. To build the OpenAPI spec and
test the MCP server against realistic payloads, we capture authenticated
browser traffic and turn it into scrubbed JSON fixtures.

This document describes that pipeline.

## Tools

- `mitmproxy` (`mitmweb` / `mitmdump`) — capture proxy.
- `python` 3.10+ — runs the capture and scrub scripts.
- `curl`, `jq` — used by the ping script.
- A logged-in browser pointed at https://app.hackthebox.com.

Install on Debian / Ubuntu:

```sh
sudo apt install mitmproxy jq python3
```

## Step 0 — Validate the token

Before any capture work, verify the API token in `.env` actually works.

```sh
./scripts/capture/ping.sh
```

Expected: `GET https://labs.hackthebox.com/api/v4/user/info -> 200` and a
short scrubbed summary line. `401` or `403` means the token is missing,
expired, or revoked — regenerate it at
https://app.hackthebox.com/profile/settings.

## Step 1 — Run mitmproxy

Start `mitmdump` with the capture addon loaded:

```sh
mitmdump -s scripts/capture/mitm_capture.py
```

Configure your browser to use `127.0.0.1:8080` as an HTTP/HTTPS proxy,
trust the `mitmproxy` CA on first run, and log in to
https://app.hackthebox.com.

Every request to `labs.hackthebox.com/api/v4/...` is written to
`scripts/capture/raw/<METHOD>__<path>__<timestamp>.json` with the full
request and response (headers + body).

The `raw/` directory is `.gitignored` because raw captures contain:

- Your bearer token in `request.headers.Authorization`.
- Your username, email, team, and personal IDs in response bodies.
- Session cookies in headers.

**Never commit raw/.**

## Step 2 — Drive the endpoint you want to document

In the browser, perform the action whose API call you want to record:
list machines, open a machine info page, submit a flag on a test box,
etc. Refresh once to catch any cached responses.

`mitmdump` will print one line per captured response. Filter to just the
endpoints you care about and note the filenames.

## Step 3 — Scrub and promote to a fixture

Once you have a raw capture, scrub it and promote it to
`scripts/capture/fixtures/<domain>/`:

```sh
python scripts/capture/scrub.py --domain machines \
  scripts/capture/raw/GET__api_v4_machine_list__1700000000000.json
```

The scrubber:

- Replaces `Authorization` / `Cookie` / `Set-Cookie` headers with `<redacted>`.
- Replaces strings that look like emails, JWTs, or long opaque tokens with
  placeholders.
- Replaces top-level `name` / `email` / `username` / `team_name` fields
  with stable example values (`example_user`, `user@example.com`, etc).
- Leaves numeric resource IDs alone — they are usually public (machine
  IDs, challenge IDs). If an ID is uniquely yours, edit the fixture by
  hand after scrubbing.

After scrubbing, **open the file and read it.** Confirm nothing identifies
you. The scrubber is conservative but not infallible.

## Step 4 — Use the fixture

- Cite the fixture as the source example in `openapi/openapi.yaml`. Each
  endpoint should have at least one example sourced from a real capture.
- Use the fixture in unit tests via `httptest` so client code is exercised
  against real-shape payloads instead of hand-written mocks.

## Folder layout

```
scripts/capture/
  ping.sh                 # read-only auth check, no on-disk output
  mitm_capture.py         # mitmproxy addon, writes to raw/
  scrub.py                # raw/<file>.json -> fixtures/<domain>/<file>.json
  raw/                    # gitignored, contains tokens + PII
  fixtures/
    .gitkeep
    machines/             # one file per captured endpoint response
    challenges/
    ...
```

## Troubleshooting

- **`mitmdump` can't intercept HTTPS.** Install the `mitmproxy` CA in your
  browser. See https://docs.mitmproxy.org/stable/concepts-certificates/.
- **No files written to `raw/`.** The addon only records calls to
  `labs.hackthebox.com/api/v4/`. Other hosts are ignored. Override with
  `HTB_CAPTURE_HOST` / `HTB_CAPTURE_PREFIX` env vars if HTB changes the API
  base.
- **`ping.sh` returns 401 immediately.** Token expired or was revoked.
  Generate a new one in HTB profile settings.
