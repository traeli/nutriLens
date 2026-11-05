#!/usr/bin/env python3
import argparse
import json
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("snapshot")
    parser.add_argument("--phase", choices=("baseline", "modified", "rollback"), required=True)
    args = parser.parse_args()
    data = json.load(open(args.snapshot, encoding="utf-8"))
    cities = {row["code"] for row in data["cities"]}
    places = data["places"]
    routes = data["city_routes"]
    stops = data["city_route_stops"]
    if args.phase == "baseline":
        assert len(places) == 26 and len(routes) == 5 and len(stops) == 4
        print("BASELINE PASS cities=6 places=26 routes=5 stops=4")
        return
    if args.phase == "modified":
        assert {"430100", "440300"} <= cities
        assert len(places) == 34 and len(routes) == 7 and len(stops) == 24
        for city in ("310000", "320100", "320500", "330100", "430100", "440300"):
            assert any(row["city_code"] == city and row["is_featured"] and row["enabled"] for row in routes)
        assert all(float(row["longitude"]) and float(row["latitude"]) for row in places if row["city_code"] in {"430100", "440300"} and row["status"] == "active")
        place_ids = {str(row["id"]) for row in places}
        route_ids = {str(row["id"]) for row in routes}
        assert all(str(row["place_id"]) in place_ids and str(row["route_id"]) in route_ids for row in stops)
        assert len({(str(row["route_id"]), int(row["sort_order"])) for row in stops}) == len(stops)
        print("MODIFIED PASS cities=6 places=34 routes=7 stops=24 featured_cities=6")
        return
    assert len(places) == 26 and len(routes) == 5 and len(stops) == 4
    print("ROLLBACK PASS cities=6 places=26 routes=5 stops=4 restored=true")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"FAIL {exc}", file=sys.stderr)
        raise
