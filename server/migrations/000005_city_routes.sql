BEGIN;

CREATE TABLE IF NOT EXISTS city_routes (
    id BIGSERIAL PRIMARY KEY,
    city_code VARCHAR(16) NOT NULL REFERENCES cities(code) ON DELETE CASCADE,
    title VARCHAR(120) NOT NULL,
    tag VARCHAR(32) NOT NULL DEFAULT '',
    meta VARCHAR(120) NOT NULL DEFAULT '',
    keyword VARCHAR(64) NOT NULL DEFAULT '',
    image VARCHAR(1024) NOT NULL DEFAULT '',
    footnote VARCHAR(240) NOT NULL DEFAULT '',
    stops JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (jsonb_typeof(stops) = 'array')
);

CREATE INDEX IF NOT EXISTS idx_city_routes_city ON city_routes(city_code, enabled, sort_order);
CREATE UNIQUE INDEX IF NOT EXISTS idx_city_routes_city_title ON city_routes(city_code, title);

COMMENT ON TABLE city_routes IS 'Operator-managed city dining routes rendered on the mini program home page';
COMMENT ON COLUMN city_routes.stops IS 'Ordered stop names, e.g. ["老周生煎","阿庆牛肉面"]';
COMMENT ON COLUMN city_routes.keyword IS 'Search keyword used when the route card is tapped';

INSERT INTO city_routes (city_code, title, tag, meta, keyword, image, footnote, stops, is_featured, enabled, sort_order) VALUES
    (
        '310000', '曹家渡的四站早餐', '早餐路线', '2.5小时 · 步行2.8km', '曹家渡',
        '/static/dining/noodle-shop.jpg', '从热气腾腾，到一杯咖啡',
        $json$["老周生煎","阿庆牛肉面","弄堂豆浆","Mono Coffee"]$json$::jsonb,
        TRUE, TRUE, 10
    ),
    (
        '310000', '下班后还亮着灯的社区小馆', '小巷路线', '静安 · 晚餐 · 6 家', '南京西路',
        '/static/dining/corner-cafe.jpg', '',
        $json$[]$json$::jsonb,
        FALSE, TRUE, 20
    ),
    (
        '330100', '从西湖边拐进大井巷', '小巷路线', '1.8小时 · 步行1.6km', '大井巷',
        '/static/dining/corner-cafe.jpg', '',
        $json$[]$json$::jsonb,
        FALSE, TRUE, 10
    ),
    (
        '320500', '从平江路拐进寻常巷陌', '小巷路线', '1.8小时 · 步行1.6km', '平江路',
        '/static/dining/noodle-detail.jpg', '',
        $json$[]$json$::jsonb,
        FALSE, TRUE, 10
    ),
    (
        '320100', '从一碗面开始认识老门东', '早餐路线', '2小时 · 步行2.2km', '老门东',
        '/static/dining/noodle-shop.jpg', '',
        $json$[]$json$::jsonb,
        FALSE, TRUE, 10
    )
ON CONFLICT (city_code, title) DO NOTHING;

COMMIT;
