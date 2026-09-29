BEGIN;

CREATE TABLE IF NOT EXISTS city_poster_themes (
    city_code VARCHAR(16) PRIMARY KEY REFERENCES cities(code) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    title VARCHAR(80) NOT NULL,
    subtitle VARCHAR(180) NOT NULL DEFAULT '',
    background_image VARCHAR(1024) NOT NULL DEFAULT '',
    accent_color VARCHAR(16) NOT NULL DEFAULT '#C7FF35',
    secondary_color VARCHAR(16) NOT NULL DEFAULT '#F2E7C9',
    motifs JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (jsonb_typeof(motifs) = 'array')
);

COMMENT ON TABLE city_poster_themes IS 'Non-geographic visual themes for personal dining-memory posters';
COMMENT ON COLUMN city_poster_themes.motifs IS 'Decorative words only; never stores boundaries, roads, directions, or coordinates';

INSERT INTO city_poster_themes (
    city_code, version, title, subtitle, background_image,
    accent_color, secondary_color, motifs
) VALUES
    ('310000', 1, '上海食光星图', '梧桐影里，记下认真吃饭的夜晚', '/static/posters/food-memory-night-v1.jpg', '#C7FF35', '#F1E5C8', '["梧桐","弄堂","夜色"]'::jsonb),
    ('330100', 1, '杭州食光星图', '风经过湖面，也经过一餐一饭', '/static/posters/food-memory-night-v1.jpg', '#BDEB7D', '#F2E7C9', '["桂花","晚风","茶香"]'::jsonb),
    ('320500', 1, '苏州食光星图', '把巷口与热汤，收进生活的星河', '/static/posters/food-memory-night-v1.jpg', '#D5F56A', '#E9DDC3', '["小巷","水声","热汤"]'::jsonb),
    ('320100', 1, '南京食光星图', '晚风、灯火，以及记得住的味道', '/static/posters/food-memory-night-v1.jpg', '#D7FF4A', '#EED9BE', '["城墙","晚风","灯火"]'::jsonb)
ON CONFLICT (city_code) DO UPDATE SET
    version = EXCLUDED.version,
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    background_image = EXCLUDED.background_image,
    accent_color = EXCLUDED.accent_color,
    secondary_color = EXCLUDED.secondary_color,
    motifs = EXCLUDED.motifs,
    updated_at = NOW();

COMMIT;
