#!/usr/bin/env python3
"""Restore places.theme_code values captured before reclassification."""
from __future__ import annotations

import json
import os
from pathlib import Path
import sys

COLLECTOR = Path("/Users/Zhuanz1/develop/src/food-route-collector")
sys.path.insert(0, str(COLLECTOR / "src"))

from food_route_collector.directus import DirectusClient  # noqa: E402

SNAPSHOT = Path(__file__).with_name("places_before.json")


def main() -> int:
    before = json.loads(SNAPSHOT.read_text())
    api = DirectusClient(os.environ["DIRECTUS_URL"], os.environ["DIRECTUS_TOKEN"], timeout=30)
    current = api._request("GET", "/items/places?fields=id,theme_code&limit=-1").get("data", [])
    current_by_id = {row["id"]: row.get("theme_code") for row in current}
    changes = [(row["id"], row.get("theme_code") or "general") for row in before if current_by_id.get(row["id"]) != row.get("theme_code")]
    for row_id, theme_code in changes:
        api.update("places", row_id, {"theme_code": theme_code})
    print(f"rollback: restored={len(changes)} total={len(before)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
