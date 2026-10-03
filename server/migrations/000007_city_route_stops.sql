BEGIN;

CREATE TABLE IF NOT EXISTS city_route_stops (
    id BIGSERIAL PRIMARY KEY,
    route_id BIGINT NOT NULL REFERENCES city_routes(id) ON DELETE CASCADE,
    place_id BIGINT NOT NULL REFERENCES places(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    note VARCHAR(240) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(route_id, sort_order)
);

CREATE INDEX IF NOT EXISTS idx_city_route_stops_route ON city_route_stops(route_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_city_route_stops_place ON city_route_stops(place_id);

INSERT INTO city_route_stops(route_id, place_id, sort_order)
SELECT route.id, place.id, stop.ordinality::integer
FROM city_routes AS route
CROSS JOIN LATERAL jsonb_array_elements_text(route.stops) WITH ORDINALITY AS stop(name, ordinality)
JOIN places AS place ON place.city_code = route.city_code AND LOWER(place.name) = LOWER(stop.name)
WHERE NOT EXISTS (
    SELECT 1 FROM city_route_stops existing
    WHERE existing.route_id = route.id AND existing.sort_order = stop.ordinality::integer
)
ON CONFLICT (route_id, sort_order) DO NOTHING;

COMMENT ON TABLE city_route_stops IS '轻量餐饮攻略路线中按顺序排列的标准地点';
COMMENT ON COLUMN city_route_stops.note IS '站点的可选攻略文案，实际导航交由用户地图应用完成';

COMMIT;
