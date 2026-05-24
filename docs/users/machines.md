# Machines domain

The `machines` domain exposes HackTheBox box-related endpoints as MCP
tools. Enable it with:

```sh
htb-app-mcp --enable=machines
```

You can combine domains, e.g. `--enable=machines,challenges` once
challenges lands.

## Tools

This is the first machines PR. It covers the full read-only surface we
captured: spawn-status, recommendations, info, walkthroughs, reviews,
guided/adventure progression, official writeup download, and a couple
of small lookups. Lifecycle (spawn / stop / extend) and flag
submission ship in follow-up PRs once those endpoints are captured.

### `machines_get_active_spawn`

Returns the caller's currently-spawned machine, or text saying no machine
is active.

| Argument | Type | Required | Notes |
|----------|------|----------|-------|
| _none_   |      |          |       |

Example: "What machine am I currently running?"

### `machines_get_recommended`

Returns the two recommendation slots shown on the HTB home page (e.g. one
seasonal box, one staff pick) plus a `state[]` describing the categories
of each card.

| Argument | Type | Required | Notes |
|----------|------|----------|-------|
| _none_   |      |          |       |

Example: "What are HTB's recommended machines for me right now?"

### `machines_get_info`

Returns full info for a single machine by its display name.

| Argument | Type   | Required | Notes |
|----------|--------|----------|-------|
| `name`   | string | yes      | Case-sensitive, e.g. `Cap`. |

The response includes the canonical numeric `id` — pass it to
`machines_get_walkthroughs` and `machines_list_reviews`.

Example: "Show me everything about machine Cap."

### `machines_get_walkthroughs`

Returns the official PDF walkthrough metadata, the official video
metadata, and a list of community-authored writeup links for a machine.

| Argument     | Type    | Required | Notes |
|--------------|---------|----------|-------|
| `machine_id` | integer | yes      | Numeric id from `machines_get_info`. |

The official PDF itself is not returned by this tool (it is a separate
non-JSON endpoint and will be exposed in a later PR).

Example: "Get me the writeups for machine 351."

### `machines_list_reviews`

Returns one page of user reviews for a machine. Uses the standard
Laravel paginator envelope.

| Argument     | Type     | Required | Notes |
|--------------|----------|----------|-------|
| `machine_id` | integer  | yes      | Numeric id from `machines_get_info`. |
| `page`       | integer  | no       | 1-based page number. Default 1. |
| `per_page`   | integer  | no       | Results per page. Default 15. |
| `sort_type`  | string   | no       | `asc` or `desc`. Default `desc`. |
| `sort_by`    | string[] | no       | Sort fields. Observed value: `created_at`. |

Example: "Show me the latest 20 reviews of machine 351."

### `machines_list_walkthrough_languages`

Returns the enum of languages community writeups can be tagged with.

| Argument | Type | Required | Notes |
|----------|------|----------|-------|
| _none_   |      |          |       |

### `machines_get_random_walkthrough_machine`

Returns a random machine reference (`id`, `name`, `avatar`) that has
community walkthroughs. Typically followed by `machines_get_walkthroughs`.

| Argument | Type | Required | Notes |
|----------|------|----------|-------|
| _none_   |      |          |       |

### `machines_get_graph_matrix`

Returns the difficulty radar matrix for a machine. Three score blocks
(aggregate / maker / user) across five skill axes (ctf, custom, cve,
enum, real).

| Argument     | Type    | Required | Notes |
|--------------|---------|----------|-------|
| `machine_id` | integer | yes      | Numeric id from `machines_get_info`. |

### `machines_list_tasks`

Returns guided-mode tasks for a machine. Each task chains via
`prerequisite_id`.

| Argument     | Type    | Required | Notes |
|--------------|---------|----------|-------|
| `machine_id` | integer | yes      | Numeric id from `machines_get_info`. |

**Sensitivity:** the `flag` field contains the actual plaintext flag
value for any task the caller has already completed. Treat the response
as user-progress data — do not paste it into shared chats or logs.

### `machines_list_adventure_steps`

Returns adventure-mode steps for a machine (typically `Submit User Flag`
and `Submit Root Flag`). The `flag` field on completed steps is a
textual indicator (e.g. `"User flag owned"`), not the flag value.

| Argument     | Type    | Required | Notes |
|--------------|---------|----------|-------|
| `machine_id` | integer | yes      | Numeric id from `machines_get_info`. |

### `machines_save_official_writeup_pdf`

Downloads the official PDF writeup and saves it to a directory. Returns
the saved path and byte count. Useful for bulk-downloading a writeup
library locally.

| Argument     | Type    | Required | Notes |
|--------------|---------|----------|-------|
| `machine_id` | integer | yes      | Numeric id from `machines_get_info`. |
| `output_dir` | string  | yes      | Absolute path to an existing writable directory. |
| `filename`   | string  | no       | Defaults to `machine_<id>.pdf`. |

Example: "Save the writeup for machine 351 to ~/htb-writeups/."

## What is intentionally not here

- **Spawn / stop / extend / reset.** Lifecycle is a separate cluster
  PR, gated on capturing the lifecycle endpoints (`/machine/play/{id}`,
  `/machine/terminate`, ...).
- **Flag submission.** Same — separate cluster.
- **General "list all machines"** (active / retired / by-OS / search).
  Not captured in this PR. Tracked in `docs/api/endpoint-checklist.md`.
