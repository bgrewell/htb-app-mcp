# Machines domain

The `machines` domain exposes HackTheBox box-related endpoints as MCP
tools. Enable it with:

```sh
htb-app-mcp --enable=machines
```

You can combine domains, e.g. `--enable=machines,challenges` once
challenges lands.

## Tools

This is the cluster shipped in the first machines PR. Lifecycle
(spawn / stop / extend) and flag submission ship in follow-up PRs.

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

## What is intentionally not here

- **Spawn / stop / extend / reset.** Lifecycle is a separate cluster
  PR, gated on capturing the lifecycle endpoints (`/machine/play/{id}`,
  `/machine/terminate`, ...).
- **Flag submission.** Same — separate cluster.
- **The official PDF writeup.** `/machine/writeup/{id}` returns a 1+ MB
  PDF and needs a different response handling strategy. Tracked in
  `docs/api/endpoint-checklist.md`.
- **General "list all machines"** (active / retired / by-OS / search).
  Not captured in this PR. Tracked in the checklist.
