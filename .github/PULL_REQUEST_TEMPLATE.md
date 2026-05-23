<!--
PRs cluster a logical sub-group of issues from a single phase.
Not one PR per issue. Not one PR spanning multiple phases.
-->

## Summary

<!-- One paragraph: what this PR does and why. -->

## Phase / domain

- Phase: <!-- e.g. 2 — Machines -->
- Domain: <!-- e.g. machines, or "infra" / "docs" for cross-cutting -->

## Closes

<!-- One line per issue this PR closes. -->
- Closes #
- Closes #

## Checklist

- [ ] CI green (build, vet, test, golangci-lint, gitleaks, redocly)
- [ ] `openapi/openapi.yaml` updated if endpoints changed; `redocly lint` clean
- [ ] User docs updated under `docs/users/<domain>.md` for any new tool
- [ ] Active tool count updated in `docs/developers/tool-count.md` if tools added/removed
- [ ] No tokens, usernames, or personal identifiers in fixtures or examples
- [ ] Copilot review requested and comments addressed
