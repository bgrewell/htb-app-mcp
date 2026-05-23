# OpenAPI authoring conventions

`openapi/openapi.yaml` is the canonical description of the HackTheBox
main-app API as we understand it. The MCP server is generated from
manual code that mirrors this spec — if the two disagree, the spec wins
and the client gets fixed.

These conventions exist so the spec stays internally consistent across
many phases and many contributors.

## File layout

- One file: `openapi/openapi.yaml`.
- If a single domain's schemas grow past ~300 lines, split them into
  `openapi/components/<domain>.yaml` and `$ref` them from the main file.
- Do not split paths into separate files. Path entries are short and
  benefit from being together for cross-cutting greps.

## Versioning

- `info.version` follows the htb-app-mcp release version, not the HTB
  API version. The HTB API is `v4` (encoded in `servers[0].url`).
- Bump `info.version` only in release PRs, not per-endpoint.

## Tags

- One tag per HackTheBox domain, names taken from this list:
  `machines`, `challenges`, `sherlocks`, `profile`, `rankings`,
  `tracks`, `pro-labs`, `fortresses`, `seasons`, `vpn`, `search`.
- Every operation has exactly one tag.
- Do not invent new tags without updating this doc.

## Operations

- `operationId` is `camelCase` and matches the Go client method name
  exactly. Examples: `listMachines`, `getMachineInfo`, `spawnMachine`,
  `submitMachineFlag`.
- Operation `summary` is one short sentence, sentence case, no trailing
  period.
- `description` is optional but encouraged for non-obvious behavior
  (e.g. "Returns 403 when the caller has not solved the prerequisite
  machine").
- Every operation specifies `security: [{bearerAuth: []}]` unless the
  endpoint is genuinely unauthenticated (rare).

## Request shapes

- Query parameters use `snake_case` to match HTB's actual API.
- Path parameters use `{snake_case}`.
- Request bodies are `application/json` unless the live API uses a
  different content type (capture-driven).
- Every parameter has a `description` and, where the set is bounded, an
  `enum`.

## Response shapes

- Every successful response has at least one `example` keyed by an
  HTTP status (`'200'`).
- Examples are sourced from `scripts/capture/fixtures/<domain>/`. Cite
  the fixture filename in a comment near the example.
- Errors reuse the shared responses in `components.responses`
  (`Unauthorized`, `NotFound`, `RateLimited`). Define a domain-specific
  response only when the shape diverges from the common `Error`.
- Paginated list endpoints reference `components.schemas.Pagination`
  for the envelope metadata and inline a typed `items: array` for the
  data.

## Schemas

- Schema names are `PascalCase`. Domain prefix when the name would
  otherwise collide (`MachineReview`, not `Review`).
- Required fields go in a `required:` list, not via `nullable: false`.
- Date-times use `format: date-time` (RFC 3339). Date-only fields use
  `format: date`. Timestamps that HTB returns as Unix seconds get
  `format: int64` with a `description` noting the unit.
- Numeric IDs are `integer`, `format: int64`, even when they currently
  fit in int32. HTB has not promised stability on width.

## Linting

`redocly lint openapi/openapi.yaml` must pass before commit. CI runs it
on every PR. Configuration lives in `.redocly.yaml` (added when the
first lint warning needs taming).

## Anti-patterns

- Do not document an endpoint you have not personally captured or
  cross-referenced against the Propolisa Postman collection.
- Do not infer response fields. If a field appears in the captured
  payload, document it; if not, leave it out.
- Do not use `anyOf` / `oneOf` to paper over response variation without
  evidence from at least two distinct captures of the divergent
  responses.
- Do not commit examples that contain identifying values — see
  `docs/api/capture.md` for the scrub workflow.
