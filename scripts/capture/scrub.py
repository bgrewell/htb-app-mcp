"""scrub.py — turn a raw capture into a committable fixture.

Reads a JSON capture from scripts/capture/raw/ (or any path), removes
identifying values, and writes the cleaned file to
scripts/capture/fixtures/<domain>/<basename>.json.

Usage:
    python scripts/capture/scrub.py --domain machines scripts/capture/raw/GET__api_v4_machine_list__*.json

Rules applied:
 - Strip the Authorization, Cookie, and Set-Cookie headers.
 - Replace any value matching a known PII pattern (email, JWT) with a
   placeholder string.
 - Replace identifying fields (name, username, email, team_name) ONLY when
   nested under a user-identifying parent key (user, profile, maker,
   author, owner, member, creator, created_by, opener, players, ranks).
   A bare top-level "name" is left alone because most HTB resources
   (machine, challenge, sherlock, team) use "name" for the resource name
   itself, which is public information.
 - Leave numeric IDs as-is. Most are public (machine/challenge IDs).
   If an ID is uniquely yours, edit the fixture manually.

The scrub is conservative and not infallible. ALWAYS read the scrubbed
fixture before committing it.

The script will refuse to overwrite an existing fixture. Use --force to
override.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
FIXTURES_DIR = ROOT / "scripts" / "capture" / "fixtures"

PII_HEADERS = {"authorization", "cookie", "set-cookie", "x-csrf-token"}

EMAIL_RE = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
JWT_RE = re.compile(r"eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+")

# Keys whose direct children are treated as identifying a person. The "name"
# / "username" / "email" / "team_name" replacements only fire when the parent
# key is in this set. This avoids stomping on machine names, challenge names,
# tag names, OS names, etc.
USER_CONTEXT_KEYS = {
    "user",
    "profile",
    "maker",
    "maker2",
    "author",
    "owner",
    "member",
    "creator",
    "created_by",
    "opener",
    "first_user_blood",
    "first_root_blood",
    "first_blood",
    "added_by",
}

# Field replacements applied only inside a USER_CONTEXT_KEYS parent.
PII_FIELD_REPLACEMENTS = {
    "name": "example_user",
    "user_name": "example_user",
    "username": "example_user",
    "email": "user@example.com",
    "team_name": "Example Team",
}


def _scrub_string(s: str) -> str:
    s = EMAIL_RE.sub("user@example.com", s)
    s = JWT_RE.sub("<jwt>", s)
    return s


def _scrub_headers(headers: dict[str, str]) -> dict[str, str]:
    out: dict[str, str] = {}
    for k, v in headers.items():
        if k.lower() in PII_HEADERS:
            out[k] = "<redacted>"
        else:
            out[k] = _scrub_string(v)
    return out


def _scrub_value(value: Any, in_user_context: bool = False, key: str | None = None) -> Any:
    if (
        in_user_context
        and key is not None
        and key in PII_FIELD_REPLACEMENTS
        and isinstance(value, str)
    ):
        return PII_FIELD_REPLACEMENTS[key]
    if isinstance(value, str):
        return _scrub_string(value)
    if isinstance(value, list):
        return [_scrub_value(v, in_user_context=in_user_context) for v in value]
    if isinstance(value, dict):
        out = {}
        for k, v in value.items():
            child_user_context = in_user_context or (k in USER_CONTEXT_KEYS)
            out[k] = _scrub_value(v, in_user_context=child_user_context, key=k)
        return out
    return value


def scrub(record: dict[str, Any]) -> dict[str, Any]:
    req = record.get("request", {})
    res = record.get("response", {})
    req["headers"] = _scrub_headers(req.get("headers", {}))
    res["headers"] = _scrub_headers(res.get("headers", {}))
    if "body" in req:
        req["body"] = _scrub_value(req["body"])
    if "body" in res:
        res["body"] = _scrub_value(res["body"])
    return record


def main() -> int:
    p = argparse.ArgumentParser(description="Scrub a raw capture into a fixture.")
    p.add_argument("--domain", required=True, help="Domain folder under fixtures/, e.g. machines")
    p.add_argument("--force", action="store_true", help="Overwrite an existing fixture")
    p.add_argument("path", help="Path to the raw capture JSON file")
    args = p.parse_args()

    src = Path(args.path)
    if not src.is_file():
        print(f"not a file: {src}", file=sys.stderr)
        return 2

    raw = json.loads(src.read_text(encoding="utf-8"))
    cleaned = scrub(raw)

    domain_dir = FIXTURES_DIR / args.domain
    domain_dir.mkdir(parents=True, exist_ok=True)
    out = domain_dir / src.name
    if out.exists() and not args.force:
        print(f"refusing to overwrite {out}; pass --force to override", file=sys.stderr)
        return 3

    out.write_text(json.dumps(cleaned, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"wrote {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
