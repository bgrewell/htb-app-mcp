"""mitm_capture.py — record HTB main-app API traffic into per-endpoint JSON.

Load this script into mitmproxy / mitmweb / mitmdump:

    mitmdump -s scripts/capture/mitm_capture.py

Then route your browser through the proxy (default :8080) and log in to
https://app.hackthebox.com. Every request to the API base path is recorded
to scripts/capture/raw/<METHOD>__<path>__<timestamp>.json.

Raw captures contain personal data (your username, team, IDs, bearer
token in request headers). They live under scripts/capture/raw/ which is
gitignored. Run scripts/capture/scrub.py before promoting anything into
scripts/capture/fixtures/.
"""

from __future__ import annotations

import json
import os
import time
from pathlib import Path
from typing import Any

from mitmproxy import http  # type: ignore[import-not-found]

API_HOST = os.environ.get("HTB_CAPTURE_HOST", "labs.hackthebox.com")
API_PREFIX = os.environ.get("HTB_CAPTURE_PREFIX", "/api/v4/")

ROOT = Path(__file__).resolve().parents[2]
RAW_DIR = ROOT / "scripts" / "capture" / "raw"
RAW_DIR.mkdir(parents=True, exist_ok=True)


def _sanitize_path(path: str) -> str:
    return path.strip("/").replace("/", "_").replace("?", "__").replace("&", "_") or "root"


def _body_to_jsonable(content: bytes | None) -> Any:
    if not content:
        return None
    try:
        return json.loads(content.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError):
        return {"_non_json": True, "len": len(content)}


def response(flow: http.HTTPFlow) -> None:
    if flow.request.pretty_host != API_HOST:
        return
    if not flow.request.path.startswith(API_PREFIX):
        return

    ts = int(time.time() * 1000)
    method = flow.request.method.upper()
    fname = f"{method}__{_sanitize_path(flow.request.path)}__{ts}.json"
    out = RAW_DIR / fname

    record = {
        "request": {
            "method": method,
            "url": flow.request.pretty_url,
            "path": flow.request.path,
            "headers": dict(flow.request.headers),
            "body": _body_to_jsonable(flow.request.content),
        },
        "response": {
            "status": flow.response.status_code,
            "headers": dict(flow.response.headers),
            "body": _body_to_jsonable(flow.response.content),
        },
    }
    out.write_text(json.dumps(record, indent=2, sort_keys=True), encoding="utf-8")
