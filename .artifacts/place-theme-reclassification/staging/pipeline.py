from __future__ import annotations

import json
from datetime import date
from pathlib import Path

from .directus import DirectusClient
from .geo import nearest_neighbor, route_distance, split_routes
from .models import normalize_category
from .storage import Store
from .theme_classifier import infer_theme_code


ROUTE_PROFILES = (
    ("2h", "solo", "一个人吃"), ("2h", "budget_friendly", "省心实惠"),
    ("half", "first_visit", "第一次来"), ("half", "family", "亲子同行"),
    ("day", "local_explore", "本地人探店"), ("day", "late_night", "深夜到访"),
)


def build_routes(store: Store, city_code: str, city_name: str = "城市") -> list[dict]:
    places = store.places(city_code)
    coord = lambda p: (float(p["longitude"]), float(p["latitude"]))
    routes=[]
    targets = {"2h":3,"half":4,"day":6}
    for index, (budget, scene, label) in enumerate(ROUTE_PROFILES, 1):
        eligible = [p for p in places if budget in json.loads(p.get("time_budgets_json") or "[]") or scene in json.loads(p.get("scene_codes_json") or "[]")]
        pool = eligible + [p for p in places if p not in eligible]
        if len(pool) < 3:
            continue
        offset = (index - 1) % len(pool)
        ordered = nearest_neighbor(pool[offset:] + pool[:offset], coord)
        groups = split_routes(ordered, coord, max_stops=targets[budget])
        group = next((item for item in groups if len(item) >= 3), [])
        if not group:
            continue
        routes.append({"city_code":city_code,"title":f"{city_name}·{label}{'半日路线' if budget == 'half' else '路线'}","keyword":f"auto-{city_code}-{budget}-{scene}","tag":label,"time_budget":budget,"scene_code":scene,"sort_order":index*10,"stops":[p["id"] for p in group],"total_distance_m":round(route_distance(group,coord),1)})
    store.replace_routes(city_code, routes)
    return routes


def make_plan(store: Store) -> dict:
    places=store.places(active_only=True); routes=store.routes()
    stops=[]
    for route in routes:
        stops.extend({"route_keyword":route["keyword"],"place_id":place_id,"sort_order":order,"note":"自动生成，发布前已通过距离检查"} for order,place_id in enumerate(route["stops"],1))
    return {"version":2,"coordinate_system":"GCJ-02","places":places,"city_routes":routes,"city_route_stops":stops}


def save_plan(plan: dict, output: str) -> Path:
    path=Path(output); path.parent.mkdir(parents=True, exist_ok=True); path.write_text(json.dumps(plan,ensure_ascii=False,indent=2),encoding="utf-8"); return path


def publish(client: DirectusClient, plan: dict, *, commit: bool = False) -> dict:
    summary={"dry_run":not commit,"places":0,"places_updated":0,"city_routes":0,"city_route_stops":0}
    if not commit:
        summary.update({k:len(plan.get(k,[])) for k in ("places","city_routes","city_route_stops")}); return summary
    local_to_remote={}
    for p in plan.get("places",[]):
        existing=client.find_one("places", {"poi_provider":p.get("poi_provider"),"poi_id":p.get("poi_id")}) if p.get("poi_id") else None
        payload={k:p.get(k) for k in ("name","category","theme_code","city_code","district","business_area","address","longitude","latitude","poi_provider","poi_id","status","average_cost","typical_wait_minutes","recommendation_source","recommendation_source_url","recommendation_note","signal_confidence")}
        payload["category"] = normalize_category(str(payload.get("category") or ""))
        payload["theme_code"] = infer_theme_code(
            name=str(payload.get("name") or ""),
            category=payload["category"],
            description=str(p.get("description") or p.get("recommendation_note") or ""),
            current=str(payload.get("theme_code") or ""),
        )
        payload["meal_periods"] = json.loads(p.get("meal_periods_json") or "[]")
        if existing:
            remote=client.update("places",existing["id"],payload); summary["places_updated"]+=1
        else:
            remote=client.create("places",payload); summary["places"]+=1
        local_to_remote[p["id"]]=remote["id"]
    for r in plan.get("city_routes",[]):
        existing=client.find_one("city_routes",{"city_code":r["city_code"],"keyword":r["keyword"]})
        if existing: remote=existing
        else:
            names=[p["name"] for p in plan.get("places",[]) if p["id"] in r["stops"]]
            remote=client.create("city_routes",{"city_code":r["city_code"],"title":r["title"],"keyword":r["keyword"],"tag":r.get("tag") or "美食路线","meta":f"{r.get('time_budget','')} · 约 {r['total_distance_m']/1000:.1f} km","stops":names,"enabled":True,"is_featured":False,"sort_order":r.get("sort_order",0)}); summary["city_routes"]+=1
        for order, local_id in enumerate(r["stops"],1):
            place_id=local_to_remote[local_id]
            existing_stop=client.find_one("city_route_stops",{"route_id":remote["id"],"place_id":place_id})
            if not existing_stop:
                client.create("city_route_stops",{"route_id":remote["id"],"place_id":place_id,"sort_order":order,"note":"自动生成，发布前已通过距离检查"}); summary["city_route_stops"]+=1
    return summary
