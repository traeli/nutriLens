BEGIN;

CREATE TABLE IF NOT EXISTS cities (
    code VARCHAR(16) PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    description VARCHAR(240) NOT NULL DEFAULT '',
    image VARCHAR(1024) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cities_enabled_sort ON cities(enabled, sort_order, code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cities_single_default ON cities(is_default) WHERE is_default = TRUE;

INSERT INTO cities(code, name, description, image, enabled, is_default, sort_order) VALUES
    ('310000', '上海', '在巷子里 遇见生活', '/static/dining/shanghai-city.jpg', TRUE, TRUE, 10),
    ('330100', '杭州', '沿湖而行 吃进四季', '/static/dining/corner-cafe.jpg', TRUE, FALSE, 20),
    ('320500', '苏州', '转进小巷 尝一口江南', '/static/dining/noodle-shop.jpg', TRUE, FALSE, 30),
    ('320100', '南京', '顺着城墙 找老味道', '/static/dining/noodle-detail.jpg', TRUE, FALSE, 40)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    image = EXCLUDED.image,
    enabled = EXCLUDED.enabled,
    is_default = EXCLUDED.is_default,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

COMMIT;
