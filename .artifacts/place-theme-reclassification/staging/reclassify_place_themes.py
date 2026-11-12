#!/usr/bin/env python3
"""Reclassify Directus places that still use the general route theme."""
from __future__ import annotations

import argparse
from collections import Counter
import os

from food_route_collector.directus import DirectusClient
from food_route_collector.theme_classifier import infer_theme_code


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--commit", action="store_true")
    args = parser.parse_args()
    client = DirectusClient(os.environ["DIRECTUS_URL"], os.environ["DIRECTUS_TOKEN"], timeout=30)
    rows = client._request(
        "GET",
        "/items/places?fields=id,name,category,description,theme_code&limit=-1&sort=id",
    ).get("data", [])
    before = Counter(str(row.get("theme_code") or "general") for row in rows)
    changes = []
    for row in rows:
        theme_code = infer_theme_code(
            name=row.get("name") or "",
            category=row.get("category") or "",
            description=row.get("description") or "",
            current=row.get("theme_code") or "",
        )
        if theme_code != row.get("theme_code"):
            changes.append((row["id"], theme_code))
    after = before.copy()
    for row_id, theme_code in changes:
        original = next(row for row in rows if row["id"] == row_id)
        after[str(original.get("theme_code") or "general")] -= 1
        after[theme_code] += 1
    print("before:", dict(sorted(before.items())))
    print("after:", {key: value for key, value in sorted(after.items()) if value})
    print(f"changes={len(changes)} commit={args.commit}")
    if args.commit:
        for row_id, theme_code in changes:
            client.update("places", row_id, {"theme_code": theme_code})
        print(f"updated={len(changes)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
