BEGIN;

CREATE TABLE IF NOT EXISTS publisher_verifications (
    user_id BIGINT PRIMARY KEY REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    phone_encrypted TEXT NOT NULL,
    phone_hash VARCHAR(64) NOT NULL UNIQUE,
    verification_channel VARCHAR(32) NOT NULL DEFAULT 'wechat_phone',
    provider_reference VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'verified',
    verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS content_safety_checks (
    id BIGSERIAL PRIMARY KEY,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    provider VARCHAR(32) NOT NULL,
    check_type VARCHAR(32) NOT NULL,
    result VARCHAR(24) NOT NULL,
    risk_labels JSONB NOT NULL DEFAULT '[]'::jsonb,
    raw_response_ref TEXT NOT NULL DEFAULT '',
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_content_safety_target ON content_safety_checks(target_type, target_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS moderation_tasks (
    id BIGSERIAL PRIMARY KEY,
    task_type VARCHAR(32) NOT NULL DEFAULT 'record_publish',
    target_type VARCHAR(32) NOT NULL DEFAULT 'visit_record',
    target_id BIGINT NOT NULL,
    record_version_id BIGINT REFERENCES visit_record_versions(id),
    priority INTEGER NOT NULL DEFAULT 0,
    risk_labels JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    decision VARCHAR(24) NOT NULL DEFAULT '',
    decision_reason TEXT NOT NULL DEFAULT '',
    assignee_id VARCHAR(128),
    due_at TIMESTAMPTZ,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_moderation_tasks_queue ON moderation_tasks(status, priority DESC, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_moderation_open_record ON moderation_tasks(target_type, target_id)
    WHERE status IN ('pending', 'assigned');

CREATE TABLE IF NOT EXISTS moderation_actions (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT REFERENCES moderation_tasks(id),
    operator_id VARCHAR(128),
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    action VARCHAR(32) NOT NULL,
    before_status VARCHAR(32) NOT NULL DEFAULT '',
    after_status VARCHAR(32) NOT NULL DEFAULT '',
    reason_code VARCHAR(64) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_moderation_actions_target ON moderation_actions(target_type, target_id, created_at DESC);

CREATE TABLE IF NOT EXISTS place_favorites (
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    place_id BIGINT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, place_id)
);

CREATE TABLE IF NOT EXISTS helpful_votes (
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    record_id BIGINT NOT NULL REFERENCES visit_records(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, record_id)
);

CREATE TABLE IF NOT EXISTS outdated_signals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    record_id BIGINT NOT NULL REFERENCES visit_records(id) ON DELETE CASCADE,
    reason TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    handled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_outdated_pending_unique ON outdated_signals(user_id, record_id) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS content_reports (
    id BIGSERIAL PRIMARY KEY,
    reporter_user_id BIGINT NOT NULL REFERENCES nutrilens_users(id),
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    reason_code VARCHAR(64) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    assigned_to VARCHAR(128),
    result TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_content_reports_user ON content_reports(reporter_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_content_reports_queue ON content_reports(status, priority DESC, created_at);

CREATE TABLE IF NOT EXISTS appeal_cases (
    id BIGSERIAL PRIMARY KEY,
    case_type VARCHAR(32) NOT NULL DEFAULT 'user_appeal',
    record_id BIGINT REFERENCES visit_records(id),
    complainant_type VARCHAR(32) NOT NULL DEFAULT 'user',
    complainant_user_id BIGINT REFERENCES nutrilens_users(id),
    contact_encrypted TEXT NOT NULL DEFAULT '',
    claim_text TEXT NOT NULL,
    query_token_hash VARCHAR(64) NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    decision TEXT NOT NULL DEFAULT '',
    assigned_to VARCHAR(128),
    response_due_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_appeal_cases_user ON appeal_cases(complainant_user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS case_materials (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES appeal_cases(id) ON DELETE CASCADE,
    submitted_by_type VARCHAR(32) NOT NULL,
    submitted_by_id BIGINT,
    material_type VARCHAR(32) NOT NULL,
    object_key VARCHAR(512) NOT NULL DEFAULT '',
    text_content TEXT NOT NULL DEFAULT '',
    access_level VARCHAR(24) NOT NULL DEFAULT 'reviewer',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS feature_flags (
    key VARCHAR(64) PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    description TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO feature_flags(key, enabled, description) VALUES
    ('public_submission_enabled', TRUE, '公开申请总开关；紧急情况下关闭'),
    ('media_upload_enabled', TRUE, '公开图片和私密消费凭证上传开关')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS route_favorites (
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    route_id BIGINT NOT NULL REFERENCES city_routes(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, route_id)
);

CREATE TABLE IF NOT EXISTS route_journeys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    route_id BIGINT NOT NULL REFERENCES city_routes(id) ON DELETE CASCADE,
    status VARCHAR(24) NOT NULL DEFAULT 'started',
    completed_stop_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_route_journeys_user ON route_journeys(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS nutrition_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES nutrilens_users(id) ON DELETE CASCADE,
    meal_period VARCHAR(24) NOT NULL DEFAULT '',
    eaten_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_type VARCHAR(24) NOT NULL DEFAULT 'manual',
    description TEXT NOT NULL DEFAULT '',
    foods JSONB NOT NULL DEFAULT '[]'::jsonb,
    calories NUMERIC(10,2),
    protein_grams NUMERIC(10,2),
    fat_grams NUMERIC(10,2),
    carbohydrate_grams NUMERIC(10,2),
    image_object_key VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_nutrition_records_user ON nutrition_records(user_id, eaten_at DESC) WHERE deleted_at IS NULL;

COMMIT;
