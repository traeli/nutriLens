#!/usr/bin/env python3
"""Snapshot, verify, or roll back the Directus place_themes icon records."""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import sys

COLLECTOR_ROOT = Path("/Users/Zhuanz1/develop/src/food-route-collector")
sys.path.insert(0, str(COLLECTOR_ROOT / "src"))

from food_route_collector.directus import DirectusClient  # noqa: E402

THEME_CODES = (
    "dumpling_bun", "noodles", "rice_meal", "local_dishes", "hotpot",
    "bbq_skewers", "street_snack", "breakfast", "coffee_tea", "dessert",
    "bakery", "seafood", "vegetarian", "bar_drinks", "general",
)
BASE_URL = "https://nutrilens-1429775067.cos.ap-guangzhou.myqcloud.com/cos-assets/static/route-theme-icons"
BACKUP = Path(__file__).with_name("directus_before.json")


def client() -> DirectusClient:
    return DirectusClient(os.environ["DIRECTUS_URL"], os.environ["DIRECTUS_TOKEN"], timeout=30)


def snapshot() -> None:
    api = client()
    rows = {code: api.find_one("place_themes", {"code": code}) for code in THEME_CODES}
    BACKUP.write_text(json.dumps(rows, ensure_ascii=False, indent=2) + "\n")
    print(f"snapshot: existing={sum(row is not None for row in rows.values())} missing={sum(row is None for row in rows.values())}")


def verify() -> None:
    api = client()
    rows = [api.find_one("place_themes", {"code": code}) for code in THEME_CODES]
    expected = {code: f"{BASE_URL}/{code}.png" for code in THEME_CODES}
    mismatches = [code for code, row in zip(THEME_CODES, rows) if row is None or row.get("icon_url") != expected[code] or row.get("enabled") is not True]
    if mismatches:
        raise SystemExit(f"verify failed: {','.join(mismatches)}")
    print("verify: rows=15 icon_urls=15 enabled=15 mismatches=0")


def rollback() -> None:
    api = client()
    before = json.loads(BACKUP.read_text())
    restored = deleted = 0
    for code in THEME_CODES:
        current = api.find_one("place_themes", {"code": code})
        original = before[code]
        if original is None:
            if current is not None:
                api._request("DELETE", f"/items/place_themes/{code}")
                deleted += 1
        elif current is None:
            api.create("place_themes", original)
            restored += 1
        else:
            api.update("place_themes", code, original)
            restored += 1
    print(f"rollback: restored={restored} deleted={deleted}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=("snapshot", "verify", "rollback"))
    args = parser.parse_args()
    {"snapshot": snapshot, "verify": verify, "rollback": rollback}[args.mode]()
