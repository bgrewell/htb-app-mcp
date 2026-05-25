# Endpoint discovery checklist

A running list of HackTheBox main-app endpoints to probe and document.
The list seeds from third-party references — Propolisa's docs and the
`noaslr/htb-mcp-server` Go source — plus endpoints we have already
observed in our own captures. The third-party references are stale on
**shapes** (verified false positive on `/machine/active` — see
`scripts/capture/fixtures/machines/list_active.json`), so they are
treated as discovery hints only. Every entry here has to be captured
and verified before it lands in `openapi/openapi.yaml`.

## Legend

| Status | Meaning |
|--------|---------|
| `captured` | We have a scrubbed fixture under `scripts/capture/fixtures/<domain>/`. |
| `documented` | The endpoint is in `openapi/openapi.yaml` with an example from a fixture. |
| `probe` | Hint from a reference; we have not captured it yet. |
| `verified-gone` | We probed and got 404/410/route-not-found. The reference is stale. |

## Machines

| Path | Method | Status | Source | Notes |
|------|--------|--------|--------|-------|
| `/machine/active` | GET | documented | propolisa, noaslr, ours | Singular: returns the user's currently-spawned machine. Propolisa's shape is wrong (claims it lists machines). |
| `/machine/recommended` | GET | documented | propolisa, ours | Returns `{card1, card2, state[]}` — two recommendation cards, not a list. |
| `/machine/profile/{name}` | GET | documented | ours | Lookup by name (not id). Returns `{info: {...}}`. |
| `/machine/walkthroughs/{id}` | GET | documented | ours | Walkthroughs for a machine. |
| `/machine/walkthroughs/language/list` | GET | documented | ours | Walkthrough language enum. |
| `/machine/walkthrough/random` | GET | documented | ours | Random machine that has walkthroughs. |
| `/machine/writeup/{id}` | GET | documented | ours | Official writeup — `application/pdf`, 1.26 MB. Exposed via `machines_save_official_writeup_pdf`. |
| `/machine/graph/matrix/{id}` | GET | documented | ours | Difficulty radar matrix (aggregate/maker/user across five axes). |
| `/review/machine/{id}/paginated` | GET | documented | ours | Reviews. Laravel paginator: `{data, meta, links, average, count}`. |
| `/machines/{id}/tasks` | GET | documented | ours | Plural-prefix route. Guided-mode tasks; flag field carries plaintext flag value once solved. |
| `/machines/{id}/adventure` | GET | documented | ours | Plural-prefix route. Canonical flag-submission progression (Submit User/Root Flag). |
| `/machine/list` | GET | probe | propolisa | Active-list per Propolisa. Has not been observed in our captures — Propolisa may be stale. Probe to confirm whether it still exists. |
| `/machine/list/retired` | GET | probe | propolisa | Retired-list per Propolisa. Same caveat. |
| `/machine/list/retired/paginated/?per_page=N` | GET | probe | noaslr | What noaslr uses for retired list today (Jul 2025). |
| `/machine/paginated/?per_page=N` | GET | probe | noaslr | What noaslr uses for general listing (Jul 2025). |
| `/machine/tags/list` | GET | probe | propolisa | Tag lookup. |
| `/machine/todo` | GET | probe | propolisa | User's "to-do" list of machines. |
| `/machine/play/{id}` | POST | probe | noaslr | Spawn a machine. Lifecycle PR. |
| `/machine/own` | POST | probe | noaslr | Submit a flag. Flags PR. |

## Challenges, Sherlocks, Profile, Rankings, Tracks, Pro Labs, Fortresses, Seasons, VPN, Search

These domains have not been probed yet. They get added to this file in
their respective phases.

## How to update this file

When you capture a new endpoint:

1. Add a row for it under the correct domain (or move an existing
   `probe` row to `captured`), citing the fixture file in the Notes
   column.

When you document an endpoint in `openapi/openapi.yaml`:

2. Flip the status from `captured` to `documented`.

When a probe returns 404/410/route-not-found:

3. Flip the status to `verified-gone` with the request that confirmed
   it.

Use this file as the source of "what to capture next" — pick the
oldest `probe` entry in the active phase's domain and drive the
capture.
