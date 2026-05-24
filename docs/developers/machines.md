# Machines — developer notes

Architecture and provenance notes for the `machines` domain.

## Layout

```
openapi/openapi.yaml                            # schemas + paths (tag: machines)
scripts/capture/fixtures/machines/              # scrubbed example responses
internal/htb/machines/machines.go               # typed Go client
internal/htb/machines/machines_test.go          # fixture-driven unit tests
internal/tools/machines/machines.go             # MCP tool registrar
docs/users/machines.md                          # user-facing tool reference
docs/developers/machines.md                     # this file
docs/api/endpoint-checklist.md                  # discovery / probe checklist
docs/developers/tool-count.md                   # active-tool budget
```

## Fixture provenance

Every operation documented in `openapi.yaml` is backed by a real captured
response under `scripts/capture/fixtures/machines/`. All six fixtures
were captured on 2026-05-23 from a logged-in browser session driven
through `mitmdump -s scripts/capture/mitm_capture.py`, scrubbed via
`scripts/capture/scrub.py`, and human-reviewed for residual PII.

| Fixture | Endpoint | Used by |
|---------|----------|---------|
| `list_active.json` | `GET /machine/active` | `getActiveSpawn`, `machines_get_active_spawn` |
| `list_recommended.json` | `GET /machine/recommended` | `getRecommendedMachines`, `machines_get_recommended` |
| `get_info_by_name.json` | `GET /machine/profile/Cap` | `getMachineInfo`, `machines_get_info` |
| `list_walkthroughs.json` | `GET /machine/walkthroughs/351` | `getMachineWalkthroughs`, `machines_get_walkthroughs` |
| `get_writeup.json` | `GET /machine/writeup/351` | not yet wired — see notes below |
| `list_reviews.json` | `GET /review/machine/351/paginated?...` | `listMachineReviews`, `machines_list_reviews` |

## Envelope conventions

HTB's API uses three different response envelopes in this domain:

| Envelope | Endpoints | Client behavior |
|----------|-----------|-----------------|
| `{info: {...}}` | active, profile/{name} | Unwrapped: methods return the inner type. |
| `{message: {...}}` | walkthroughs/{id} | Unwrapped. |
| flat (no envelope) | recommended, reviews | Returned as-is. |

The client always unwraps when an envelope is present so callers do not
have to thread the envelope name through every call site.

## Surprises uncovered during capture

- **`/machine/active` is singular.** Despite the URL it returns the
  caller's current spawn, not a list. Reflected in the tool name
  `machines_get_active_spawn`. Listing all active boxes requires a
  different (un-captured) endpoint — likely `/machine/paginated/`.
- **Profile lookup is by name, not by id.** `/machine/profile/{name}`
  takes the display name. The numeric `id` returned in the response is
  used by walkthroughs / reviews / lifecycle endpoints.
- **Mixed URL roots.** Most paths are `/machine/...` but
  `/machines/{id}/tasks` and `/machines/{id}/adventure` exist on a
  plural prefix. Likely two router generations on HTB's side. Future
  endpoints in this cluster keep whatever the live API uses.
- **Reviews has a custom envelope.** Laravel paginator (`data, meta,
  links`) plus two custom top-level fields: `average` (stars across all
  reviews) and `count` (total review count).
- **`/machine/writeup/{id}` returns `application/pdf`.** The capture
  body is non-JSON (1.26 MB). Not exposed as an MCP tool in this PR;
  needs a download-URL-plus-metadata wrapper.

## Field-required policy

Only `id` and `name` are marked `required` in schemas where they appear
across all captures we have. Everything else is left optional, even
fields that were present in our one observed response. Tighten when we
have multi-state captures (auth/anon, owned/not-owned, retired/active,
seasonal/normal, etc.).

## Adding new machines endpoints

1. Drive the endpoint in the proxied browser, scrub the resulting raw
   into `scripts/capture/fixtures/machines/`.
2. Add the path and schemas to `openapi/openapi.yaml` under the
   `machines` tag.
3. Add typed structs + a method on `*Client` in
   `internal/htb/machines/machines.go`. Keep the package one-file for
   now; split when it crosses ~600 LOC.
4. Add an `httptest`-driven unit test that decodes the new fixture.
5. Register an MCP tool in `internal/tools/machines/machines.go`.
   Reuse the `noInput` / `machineIDInput` patterns when applicable.
6. Update the row in `docs/api/endpoint-checklist.md` from `probe` /
   `captured` to `documented`.
7. Bump the tool count in `docs/developers/tool-count.md`.
8. Update `docs/users/machines.md` with the new tool's reference.
