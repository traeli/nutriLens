BEGIN;

ALTER TABLE places
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN places.description IS '由运营维护的客观地点介绍，不得复制用户评论';

COMMIT;
