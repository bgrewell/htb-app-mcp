# Active MCP tool count

MCP clients lose tool-selection accuracy above ~25 active tools. Every
change that adds or removes a tool must update this file.

| Domain | Tools | Count |
|--------|-------|-------|
| machines | `machines_get_active_spawn`, `machines_get_recommended`, `machines_get_info`, `machines_get_walkthroughs`, `machines_list_reviews`, `machines_list_walkthrough_languages`, `machines_get_random_walkthrough_machine`, `machines_get_graph_matrix`, `machines_list_tasks`, `machines_list_adventure_steps`, `machines_save_official_writeup_pdf` | 11 |
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
| **Total (all enabled)** | | **11** |

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

| Domain | Original target | Current |
|--------|----------------|---------|
| machines | 8 | 11 (over by 3) |
| challenges | 5 | — |
| sherlocks | 4 | — |
| profile + rankings + tracks combined | 4 | — |
| pro-labs + fortresses + seasons combined | 3 | — |
| vpn + search combined | 2 | — |
| **Total** | **~26** | — |

Machines is currently over the 8-tool target because the first PR
documents the full captured read-only surface (per the
"document the API in totality" principle) rather than curating to a
subset. Revisit once challenges + sherlocks ship and we can see how
the sum trends. If totals cross 25, the first consolidation candidates
are likely: collapse `list_walkthrough_languages` +
`get_random_walkthrough_machine` into one rarely-used "walkthrough
discovery" tool, or move them out of the default-enabled set.
