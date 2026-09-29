BEGIN;

CREATE TABLE IF NOT EXISTS nutrilens_users (
    id BIGSERIAL PRIMARY KEY,
    open_id VARCHAR(128) NOT NULL UNIQUE,
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    avatar_url VARCHAR(512) NOT NULL DEFAULT '',
    account_status VARCHAR(24) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_nutrilens_users_account_status ON nutrilens_users(account_status);
CREATE INDEX IF NOT EXISTS idx_nutrilens_users_deleted_at ON nutrilens_users(deleted_at);

CREATE TABLE IF NOT EXISTS privacy_agreements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    agreement_type VARCHAR(64) NOT NULL,
    version VARCHAR(32) NOT NULL,
    ip_hash VARCHAR(128) NOT NULL DEFAULT '',
    client_version VARCHAR(32) NOT NULL DEFAULT '',
    agreed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, agreement_type, version)
);

CREATE TABLE IF NOT EXISTS places (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    category VARCHAR(32) NOT NULL,
    city_code VARCHAR(16) NOT NULL,
    district VARCHAR(64) NOT NULL DEFAULT '',
    business_area VARCHAR(64) NOT NULL DEFAULT '',
    address VARCHAR(300) NOT NULL,
    longitude NUMERIC(10,7) NOT NULL,
    latitude NUMERIC(10,7) NOT NULL,
    poi_provider VARCHAR(32),
    poi_id VARCHAR(128),
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    merged_into_id BIGINT REFERENCES places(id),
    created_by BIGINT REFERENCES nutrilens_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_places_city_name ON places(city_code, name);
CREATE INDEX IF NOT EXISTS idx_places_status ON places(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_places_provider_poi ON places(poi_provider, poi_id) WHERE poi_provider IS NOT NULL AND poi_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(64) NOT NULL,
    group_name VARCHAR(64) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

INSERT INTO tags(code, name, group_name, sort_order) VALUES
    ('taste', '口味', 'experience', 10),
    ('price', '价格', 'experience', 20),
    ('service', '服务', 'experience', 30),
    ('queue', '排队', 'experience', 40),
    ('hygiene_observation', '卫生观感', 'experience', 50),
    ('promotion_mismatch', '宣传不符', 'experience', 60),
    ('other', '其他', 'experience', 99)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS visit_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    place_id BIGINT NOT NULL REFERENCES places(id),
    visibility VARCHAR(16) NOT NULL DEFAULT 'private',
    publish_status VARCHAR(32) NOT NULL DEFAULT 'draft',
    risk_level VARCHAR(16) NOT NULL DEFAULT 'low',
    current_version_id BIGINT,
    submitted_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    last_confirmed_at TIMESTAMPTZ,
    helpful_count INTEGER NOT NULL DEFAULT 0,
    outdated_count INTEGER NOT NULL DEFAULT 0,
    report_count INTEGER NOT NULL DEFAULT 0,
    create_request_key VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_visit_records_user_created ON visit_records(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_visit_records_place_status ON visit_records(place_id, publish_status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_visit_records_public_feed ON visit_records(visibility, publish_status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_visit_records_deleted_at ON visit_records(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_visit_records_idempotency ON visit_records(user_id, create_request_key) WHERE create_request_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS visit_record_versions (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES visit_records(id),
    version_no INTEGER NOT NULL,
    editor_user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    visit_date DATE NOT NULL,
    consumer_type VARCHAR(32) NOT NULL DEFAULT '',
    conclusion VARCHAR(32) NOT NULL,
    price_min NUMERIC(10,2),
    price_max NUMERIC(10,2),
    average_cost NUMERIC(10,2),
    wait_minutes INTEGER,
    meal_period VARCHAR(24) NOT NULL DEFAULT '',
    dishes JSONB NOT NULL DEFAULT '[]'::jsonb,
    content TEXT NOT NULL DEFAULT '',
    visibility VARCHAR(16) NOT NULL DEFAULT 'private',
    change_summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(record_id, version_no)
);
CREATE INDEX IF NOT EXISTS idx_visit_record_versions_record ON visit_record_versions(record_id, version_no DESC);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_visit_records_current_version') THEN
        ALTER TABLE visit_records ADD CONSTRAINT fk_visit_records_current_version FOREIGN KEY (current_version_id) REFERENCES visit_record_versions(id);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS visit_record_tag_links (
    record_version_id BIGINT NOT NULL REFERENCES visit_record_versions(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id),
    PRIMARY KEY(record_version_id, tag_id)
);

CREATE TABLE IF NOT EXISTS record_media (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES visit_records(id),
    record_version_id BIGINT NOT NULL REFERENCES visit_record_versions(id),
    object_key VARCHAR(512) NOT NULL,
    public_url VARCHAR(1024) NOT NULL DEFAULT '',
    media_type VARCHAR(24) NOT NULL,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    safety_status VARCHAR(24) NOT NULL DEFAULT 'pending',
    desensitize_status VARCHAR(24) NOT NULL DEFAULT 'pending',
    processed_object_key VARCHAR(512) NOT NULL DEFAULT '',
    face_detected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_record_media_record ON record_media(record_id, record_version_id);

CREATE TABLE IF NOT EXISTS record_evidences (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES visit_records(id),
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    evidence_type VARCHAR(32) NOT NULL,
    original_object_key VARCHAR(512) NOT NULL,
    masked_object_key VARCHAR(512) NOT NULL DEFAULT '',
    verify_status VARCHAR(24) NOT NULL DEFAULT 'pending',
    verified_by BIGINT,
    verified_at TIMESTAMPTZ,
    retention_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_record_evidences_record ON record_evidences(record_id);

CREATE TABLE IF NOT EXISTS user_trust_profiles (
    user_id BIGINT PRIMARY KEY REFERENCES nutrilens_users(id),
    trust_level VARCHAR(24) NOT NULL DEFAULT 'new',
    account_age_score INTEGER NOT NULL DEFAULT 0,
    verified_record_count INTEGER NOT NULL DEFAULT 0,
    violation_count INTEGER NOT NULL DEFAULT 0,
    daily_publish_limit INTEGER NOT NULL DEFAULT 1,
    risk_flags JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contribution_accounts (
    user_id BIGINT PRIMARY KEY REFERENCES nutrilens_users(id),
    total_points INTEGER NOT NULL DEFAULT 0,
    month_points INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS badges (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon_url VARCHAR(1024) NOT NULL DEFAULT '',
    condition_type VARCHAR(64) NOT NULL,
    condition_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_badges (
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    badge_id BIGINT NOT NULL REFERENCES badges(id),
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, badge_id)
);

COMMIT;
