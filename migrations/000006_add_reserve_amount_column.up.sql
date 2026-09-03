SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE upload_sessions ADD COLUMN IF NOT EXISTS reserve_amount bigint NOT NULL DEFAULT 0;