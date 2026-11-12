#!/usr/bin/env python3
"""Idempotently publish a collector plan to Directus in bounded bulk requests."""

from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from urllib.parse import urlencode


ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "src"))

from food_route_collector.cli import load_dotenv  # noqa: E402
from food_route_collector.directus import DirectusClient  # noqa: E402
from food_route_collector.models import normalize_category  # noqa: E402


def pages(client: DirectusClient, collection: str, fields: str) -> list[dict]:
    rows: list[dict] = []
    offset = 0
    while True:
        query = urlencode({"fields": fields, "limit": 100, "offset": offset})
        batch = client._request("GET", f"/items/{collection}?{query}").get("data", [])
        rows.extend(batch)
        if len(batch) < 100:
            return rows
        offset += len(batch)


def chunks(items: list[dict], size: int) -> list[list[dict]]:
    return [items[index:index + size] for index in range(0, len(items), size)]


def _submit_resilient(client: DirectusClient, method: str, collection: str, batch: list[dict]) -> tuple[int, int]:
    try:
        client._request(method, f"/items/{collection}", batch)
        return len(batch), 0
    except RuntimeError as exc:
        # Another idempotent publisher may have inserted a key after our snapshot.
        # Split the batch so only the conflicting row is skipped.
        if len(batch) == 1:
            if "RECORD_NOT_UNIQUE" in str(exc):
                return 0, 1
            raise
        middle = len(batch) // 2
        left = _submit_resilient(client, method, collection, batch[:middle])
        right = _submit_resilient(client, method, collection, batch[middle:])
        return left[0] + right[0], left[1] + right[1]


def bulk(client: DirectusClient, method: str, collection: str, items: list[dict], size: int) -> int:
    completed = 0
    skipped = 0
    for number, batch in enumerate(chunks(items, size), 1):
        added, duplicates = _submit_resilient(client, method, collection, batch)
        completed += added
        skipped += duplicates
        print(f"{method} {collection}: batch={number} applied={completed} duplicate_skips={skipped} total={len(items)}", flush=True)
    return completed


def place_payload(place: dict) -> dict:
    keys = (
        "name", "category", "theme_code", "city_code", "district", "business_area", "address",
        "longitude", "latitude", "poi_provider", "poi_id", "status", "average_cost",
        "typical_wait_minutes", "recommendation_source", "recommendation_source_url",
        "recommendation_note", "signal_confidence",
    )
    payload = {key: place.get(key) for key in keys}
    payload["category"] = normalize_category(str(payload.get("category") or ""))
    payload["theme_code"] = str(payload.get("theme_code") or "general")
    payload["meal_periods"] = json.loads(place.get("meal_periods_json") or "[]")
    return payload


def route_payload(route: dict, place_by_local_id: dict[str, dict]) -> dict:
    names = [place_by_local_id[place_id]["name"] for place_id in route["stops"]]
    return {
        "city_code": route["city_code"],
        "title": route["title"],
        "keyword": route["keyword"],
        "tag": route.get("tag") or "美食路线",
        "meta": f"{route.get('time_budget', '')} · 约 {route['total_distance_m'] / 1000:.1f} km",
        "stops": names,
        "enabled": True,
        "is_featured": False,
        "sort_order": route.get("sort_order", 0),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("plan")
    parser.add_argument("--commit", action="store_true")
    parser.add_argument("--artifact", default="runs/directus-commit-20261106.json")
    args = parser.parse_args()

    load_dotenv()
    plan = json.loads(Path(args.plan).read_text(encoding="utf-8"))
    client = DirectusClient(os.environ["DIRECTUS_URL"], os.environ["DIRECTUS_TOKEN"], timeout=60)
    before = {name: client.count(name) for name in ("places", "city_routes", "city_route_stops")}

    remote_places = pages(client, "places", "id,poi_provider,poi_id")
    place_key_to_id = {
        (str(row.get("poi_provider") or ""), str(row.get("poi_id") or "")): row["id"]
        for row in remote_places if row.get("poi_id")
    }
    creates: list[dict] = []
    updates: list[dict] = []
    for place in plan.get("places", []):
        payload = place_payload(place)
        remote_id = place_key_to_id.get((str(place.get("poi_provider") or ""), str(place.get("poi_id") or "")))
        if remote_id is None:
            creates.append(payload)
        else:
            updates.append({"id": remote_id, **payload})

    remote_routes = pages(client, "city_routes", "id,city_code,keyword")
    route_key_to_row = {(str(row["city_code"]), str(row["keyword"])): row for row in remote_routes}
    place_by_local_id = {place["id"]: place for place in plan.get("places", [])}
    route_creates: list[dict] = []
    route_updates: list[dict] = []
    for route in plan.get("city_routes", []):
        payload = route_payload(route, place_by_local_id)
        existing = route_key_to_row.get((str(route["city_code"]), str(route["keyword"])))
        if existing:
            route_updates.append({"id": existing["id"], **payload})
        else:
            route_creates.append(payload)

    preview = {
        "dry_run": not args.commit,
        "before": before,
        "planned": {
            "places_create": len(creates),
            "places_update": len(updates),
            "city_routes_create": len(route_creates),
            "city_routes_update": len(route_updates),
            "city_route_stops_plan": len(plan.get("city_route_stops", [])),
        },
    }
    print(json.dumps(preview, ensure_ascii=False), flush=True)
    if not args.commit:
        return 0

    # Fixed publication order: places, city_routes, city_route_stops.
    bulk(client, "POST", "places", creates, 20)
    bulk(client, "PATCH", "places", updates, 20)

    remote_places = pages(client, "places", "id,poi_provider,poi_id")
    place_key_to_id = {
        (str(row.get("poi_provider") or ""), str(row.get("poi_id") or "")): row["id"]
        for row in remote_places if row.get("poi_id")
    }
    local_to_remote = {
        place["id"]: place_key_to_id[(str(place.get("poi_provider") or ""), str(place.get("poi_id") or ""))]
        for place in plan.get("places", [])
    }

    bulk(client, "POST", "city_routes", route_creates, 20)
    bulk(client, "PATCH", "city_routes", route_updates, 20)
    remote_routes = pages(client, "city_routes", "id,city_code,keyword")
    route_key_to_id = {(str(row["city_code"]), str(row["keyword"])): row["id"] for row in remote_routes}
    existing_stops = pages(client, "city_route_stops", "id,route_id,place_id")
    stop_creates: list[dict] = []
    route_by_keyword = {route["keyword"]: route for route in plan.get("city_routes", [])}
    for stop in plan.get("city_route_stops", []):
        route = route_by_keyword[stop["route_keyword"]]
        route_id = route_key_to_id[(str(route["city_code"]), str(route["keyword"]))]
        place_id = local_to_remote[stop["place_id"]]
        payload = {"route_id": route_id, "place_id": place_id, "sort_order": stop["sort_order"], "note": stop.get("note") or "自动生成，发布前已通过距离检查"}
        stop_creates.append(payload)
    target_route_ids = {str(route_key_to_id[(str(route["city_code"]), str(route["keyword"]))]) for route in plan.get("city_routes", [])}
    # Route position has a uniqueness constraint. Replacing the target routes'
    # stops avoids transient conflicts while swapping sort_order values.
    stop_deletes = [row["id"] for row in existing_stops if str(row["route_id"]) in target_route_ids]
    bulk(client, "DELETE", "city_route_stops", stop_deletes, 40)
    bulk(client, "POST", "city_route_stops", stop_creates, 40)

    after = {name: client.count(name) for name in ("places", "city_routes", "city_route_stops")}
    result = {
        **preview,
        "dry_run": False,
        "applied": {
            "places_created": len(creates),
            "places_updated": len(updates),
            "city_routes_created": len(route_creates),
            "city_routes_updated": len(route_updates),
            "city_route_stops_deleted": len(stop_deletes),
            "city_route_stops_updated": 0,
            "city_route_stops_created": len(stop_creates),
        },
        "after": after,
    }
    artifact = Path(args.artifact)
    artifact.parent.mkdir(parents=True, exist_ok=True)
    artifact.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(result, ensure_ascii=False), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
