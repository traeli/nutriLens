BEGIN;

ALTER TABLE places
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN places.description IS 'Operator-maintained factual place introduction; never copied from a user review';

COMMIT;
