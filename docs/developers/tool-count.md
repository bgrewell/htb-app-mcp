# Active MCP tool count

MCP clients lose tool-selection accuracy above ~25 active tools. Every
change that adds or removes a tool must update this file.

| Domain | Tools | Count |
|--------|-------|-------|
| machines | `machines_get_active_spawn`, `machines_get_recommended`, `machines_get_info`, `machines_get_walkthroughs`, `machines_list_reviews` | 5 |
| challenges | _not implemented yet_ | 0 |
| sherlocks | _not implemented yet_ | 0 |
| profile | _not implemented yet_ | 0 |
| rankings | _not implemented yet_ | 0 |
| tracks | _not implemented yet_ | 0 |
| pro-labs | _not implemented yet_ | 0 |
| fortresses | _not implemented yet_ | 0 |
| seasons | _not implemented yet_ | 0 |
| vpn | _not implemented yet_ | 0 |
| search | _not implemented yet_ | 0 |
| **Total (all enabled)** | | **5** |

## Discipline

- Prefer one tool that takes a `kind` / `state` parameter over multiple
  near-duplicate tools. Example: do NOT add separate
  `machines_list_active` / `machines_list_retired` if a single
  `machines_list` with `status: active|retired` works.
- Lifecycle (spawn/stop/extend/reset) likely fits in 1–2 tools, not 4.
- Document any deviation from the discipline in a PR description.

## Per-domain budget guidance

Rough target ceilings per domain so a user enabling all domains stays
near or below 25 active tools:

| Domain | Target ceiling |
|--------|----------------|
| machines | 8 |
| challenges | 5 |
| sherlocks | 4 |
| profile + rankings + tracks combined | 4 |
| pro-labs + fortresses + seasons combined | 3 |
| vpn + search combined | 2 |
| **Total** | **~26** |

These are guidelines, not hard caps. Revisit when the first three
domains have shipped.
