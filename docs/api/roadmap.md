# Roadmap to "complete" API coverage

This file is the project's working answer to the question: *what do we still
need to do to get to 100% coverage of the HackTheBox main-app API?*

The honest framing: there is no published HTB API contract, so "complete"
means "every endpoint we discover, verified by a fresh capture and exposed
where it makes sense as an MCP tool." The Propolisa Postman collection
claims 111+ endpoints (stale; ~2021), `noaslr/htb-mcp-server` implements
~12 of those, and our 2026-05-23 single home-page capture touched 26
unique paths across 9 domain prefixes. The real number today is probably
**150-200 endpoints** across all domains.

The companion `docs/api/endpoint-checklist.md` is the per-endpoint
tracker (probe / captured / documented). This file is the higher-level
plan — what's next, why, and how to close the loop on each domain.

## Status snapshot

| Phase | Scope | State | Notes |
|------|------|------|------|
| 0 | Foundation | done | PR #1 merged |
| 1 | Recon + bootstrap | done | PR #1 merged |
| **2a** | **Machines: list + info** | **done** | **PR #13 merged — 11 endpoints, 11 tools** |
| 2b | Machines: lifecycle | not started | spawn / stop / extend / reset |
| 2c | Machines: flag + ownership | not started | submit user/root flag, own status |
| 2d | Machines: general lists + tags + changelog + todo | not started | active / retired / by-OS / search / tags / changelog / to-do |
| 3 | Challenges | not started | own domain |
| 4 | Sherlocks | not started | own domain |
| 5a | Profile + Rankings | not started | many sub-endpoints |
| 5b | Tracks + Pro Labs + Fortresses + Seasons | not started | medium effort |
| 5c | VPN + Search + cross-cutting | not started | small endpoints |
| 6 | v1.0 release | not started | GoReleaser, Docker, MCP registry |

## What we already know exists

The 2026-05-23 capture pulled 51 raw files just from loading the HTB home
page. After de-duplication, those touched these domain prefixes:

| Prefix | Count | State |
|--------|-------|-------|
| `/machine/...`, `/machines/...`, `/review/machine/...` | 11 | documented (Phase 2a) |
| `/user/...` | 5 | probe (Phase 5a) |
| `/season/...` | 2 | probe (Phase 5b) |
| `/pwnbox/...` | 2 | probe (Phase 5c) |
| `/connection/status` | 1 | probe (Phase 5c) |
| `/home/banners`, `/notices`, `/navigation/main`, `/sso/redirect`, `/tags/list` | 5 | probe (cross-cutting) |

Plus references and the plan's own enumeration cover the other six
domains we haven't probed yet.

## Per-domain plan

Each domain block lists the **endpoint clusters** we'll most likely need
to capture and document, derived from a mix of public refs, the HTB app
UI structure, and the prior-domain experience. Counts are **estimates**;
real numbers come out of the capture sessions.

### Machines (Phase 2) — 11 done, ~10-15 to go

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 2a list + info ✓ | `/machine/active`, `/recommended`, `/profile/{name}`, `/walkthroughs/{id}`, `/walkthrough/random`, `/walkthroughs/language/list`, `/writeup/{id}`, `/review/machine/{id}/paginated`, `/graph/matrix/{id}`, `/machines/{id}/tasks`, `/machines/{id}/adventure` | done |
| 2b lifecycle | `/machine/play/{id}` (spawn), spawn-arena variants, terminate/stop, extend, reset, switch server, status poll | spawn a machine in the UI; stop it; extend timer |
| 2c flag + ownership | `/machine/own` (submit flag), `/machine/changelog/{id}`, own status check, ownership matrix submission, rate matrix submission | submit a real flag; rate a machine you've owned |
| 2d general lists + lookup | `/machine/paginated` (or whatever today's "all machines" endpoint is — Propolisa says `/machine/list`, noaslr uses `/machine/paginated`; both need verification), `/machine/list/retired` variant, OS filter, search-by-name, `/machine/tags/list`, `/machine/todo`, `/machine/todo/update/{id}`, changelog | click "Machines" nav, filter active/retired, change OS filter, type in search box, open todo list |

### Challenges (Phase 3) — 0 done, ~15-25 endpoints

Public refs and noaslr both touch this. Likely cluster split:

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 3a list + info | `/challenge/categories/list`, `/challenge/list`, `/challenge/list/retired`, per-category lists, `/challenge/info/{id}` (or similar), search | click "Challenges" nav, browse categories |
| 3b lifecycle (interactive) | `/challenge/{id}/start`, stop, extend; download attached files | start a challenge that has a Docker / interactive instance |
| 3c flag + ownership | `/challenge/own`, ratings, reviews, writeup metadata, walkthroughs | submit a flag, write a review |
| 3d todo + tags | `/challenge/todo`, `/challenge/todo/update/{id}`, tag lookup | add/remove a challenge from to-do |

### Sherlocks (Phase 4) — 0 done, ~10-15 endpoints

Sherlocks are HTB's investigation scenarios. Probably mirrors challenges
structure with download-evidence + Q&A submission shapes.

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 4a list + info | sherlock listing, category listing, sherlock detail | open the Sherlocks section in the UI |
| 4b evidence | download evidence archives, start sherlock | start one |
| 4c answers | submit answers per task, get progress | solve a task |

### Profile (Phase 5a — own + others) — 0 done, ~20-30 endpoints

This is the largest domain by endpoint count thanks to per-chart
breakdowns. Propolisa enumerates: `/user/info`, `/user/anonymized/id`,
`/user/connection/status`, `/user/settings`, `/user/profile/basic/{id}`,
`/user/profile/activity/{id}`, `/user/profile/bloods/{id}`,
`/user/profile/content/{id}`, `/user/profile/chart/machines/attack/{id}`,
`/user/profile/graph/{period}/{id}` (1W, 1Y observed),
`/user/profile/progress/machines/os/{id}`,
`/user/profile/progress/challenges/{id}`,
`/user/profile/progress/endgame/{id}`,
`/user/profile/progress/fortress/{id}`,
`/user/profile/progress/prolab/{id}`,
`/user/subscriptions/management`.

That's 16+ before we even probe; the real list is probably 25-30 once we
include badges, team membership, achievement endpoints, and app-token
management.

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5a-i self | `/user/info`, `/user/settings`, `/user/connection/status`, `/user/anonymized/id`, `/user/subscriptions/management` | open profile settings |
| 5a-ii other-user view | `/user/profile/{basic,activity,bloods,content}/{id}` | visit another user's profile |
| 5a-iii charts + graphs | `/user/profile/graph/{period}/{id}`, `/user/profile/chart/machines/attack/{id}`, points-over-time | scroll a user profile page |
| 5a-iv progress | `/user/profile/progress/{machines/os,challenges,endgame,fortress,prolab}/{id}` | view another user's progress tab |
| 5a-v app tokens | list / create / revoke app tokens (the same kind of token we use to auth) | go to App Tokens settings |

### Rankings (Phase 5a-vi) — 0 done, ~5-10 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5a-vi global + team rankings | top users (overall, country, season), top teams, team profile | open Rankings page; switch country/season filters |

### Tracks (Phase 5b-i) — 0 done, ~5-10 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5b-i list + info + enroll | `/tracks/list` (or similar), track detail, enroll/unenroll, progress | open Tracks, enroll in one |

### Pro Labs (Phase 5b-ii) — 0 done, ~10 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5b-ii list + info + flags | pro lab listing, detail page, machines within, flag submission | view a Pro Lab page |

### Fortresses (Phase 5b-iii) — 0 done, ~5-10 endpoints

Same shape as Pro Labs.

### Seasons (Phase 5b-iv) — 0 done, ~10-15 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5b-iv current + history | `/season/list` (probed), `/season/{id}` detail, `/season/user/{id}/ranks` (probed), leaderboards, badges | view current season, switch to past season |

### VPN (Phase 5c-i) — 0 done, ~5-10 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5c-i server list + config + switch | list VPN servers per lab, download OVPN config, switch active server, pwnbox status (`/pwnbox/{status,usage}` probed) | open Lab Access → VPN |

### Search (Phase 5c-ii) — 0 done, ~3-5 endpoints

| Cluster | Endpoints (likely) | Capture trigger |
|---------|---------------------|-----------------|
| 5c-ii global search | `/search/fetch` (noaslr uses this), category-scoped variants | use the global search bar |

### Cross-cutting (covered alongside Phase 5c)

`/home/banners`, `/notices`, `/navigation/main`, `/tags/list`, `/sso/redirect`. These are app-shell helpers, not domain endpoints. Document but
likely no MCP tool — they're context the app uses, not user actions.

## Methodology to converge on 100%

**One-time deviation from the per-cluster PR cadence in CONTRIBUTING.md:**
the remaining API work is being landed in a **single long-lived branch**
(`api-complete`) and one mega-PR. The MCP server is the product; the API
surface is a prerequisite — chunking the prerequisite into ~24 PRs
spends a lot of process time before any of it ships to a user. We pay
for that with a harder review on one big PR; the trade-off is justified
once and not repeated.

Per-cluster recipe (still applies within the branch; just no per-cluster PR):

1. **Capture session** — drive the relevant UI actions through `mitmdump`.
2. **Scrub + read** — promote captures to `scripts/capture/fixtures/<domain>/`, human-review for PII.
3. **Document** — add operations + schemas to `openapi/openapi.yaml`.
4. **Client + tools + tests** — wire `internal/htb/<domain>` + `internal/tools/<domain>`, fixture-driven `httptest`.
5. **Docs** — extend `docs/users/<domain>.md`, `docs/developers/<domain>.md`, `docs/developers/tool-count.md`.
6. **Tracking issues** — one `type/endpoint-doc` issue per endpoint stays open until the mega-PR closes them all together.
7. **Commit on `api-complete`** with a per-cluster commit message; update `endpoint-checklist.md` and this file's status snapshot.

The mega-PR opens once Phase 5c is complete (or once we hit a natural
"good enough for v1.0" stopping point). After it merges, we revert to
the normal per-cluster PR cadence for Phase 6 release work and any
post-1.0 hardening.

When are we "done"? Three escalating gates:

- **Gate 1 (claim coverage):** every section of the HTB app UI we can navigate to is reachable via some MCP tool.
- **Gate 2 (verified coverage):** we have a multi-state capture per endpoint (auth/anon, owned/not-owned, retired/active, etc.) and the spec's `required` fields reflect that evidence.
- **Gate 3 (no known gaps):** Propolisa's collection, noaslr's source, and a fresh full-site walk all yield no endpoint that isn't already in our spec.

Gate 1 is what v1.0 (Phase 6) requires. Gate 2 is post-1.0 hardening.
Gate 3 is aspirational — the API moves underneath us.

## Phase 6 (release) — outside the per-domain loop

- `goreleaser` config, multi-platform binaries, signed releases.
- Dockerfile + GHCR publishing.
- Submit to the MCP registry (`modelcontextprotocol/registry`).
- Smoke-test job in CI that exercises one tool per domain against the live API, gated by `HTB_API_KEY` secret.
- Cut `v1.0.0` once Gate 1 is met.

## Cost estimate

Very rough, based on the Phase 2a experience (~half a day of focused work
end-to-end for 11 endpoints):

| Phase | Endpoints (est.) | PR clusters (est.) | Effort (focused half-days) |
|------|-----------------|--------------------|----------------------------|
| 2b/c/d (rest of machines) | 10-15 | 3 | 2 |
| 3 (challenges) | 15-25 | 4 | 3 |
| 4 (sherlocks) | 10-15 | 3 | 2 |
| 5a (profile + rankings) | 25-35 | 5 | 4 |
| 5b (tracks/prolabs/fortresses/seasons) | 30-40 | 6 | 5 |
| 5c (vpn/search/cross-cutting) | 10-15 | 3 | 2 |
| **Total to Gate 1** | **~110-145** | **~24** | **~18** |
| 6 (release) | — | 1 | 1 |

The bottleneck is **capture time**, not coding time — every cluster
needs the user to drive UI actions in the proxied browser. Coding +
docs + PR for one cluster runs ~30-60 minutes once captures are in
hand.

## How to use this file

- Pick the next cluster you want to ship. Read its "capture trigger" column.
- Run a capture session for just those actions.
- Hand off the raws; the per-cluster recipe takes it from there.
- Update this file's status snapshot and the per-domain block as clusters land.
