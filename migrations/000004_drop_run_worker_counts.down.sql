SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE runs ADD COLUMN IF NOT EXISTS worker_count bigint NOT NULL DEFAULT 0;
ALTER TABLE runs ADD COLUMN IF NOT EXISTS completed_worker_count bigint NOT NULL DEFAULT 0;
ALTER TABLE runs ADD COLUMN IF NOT EXISTS active_worker_count bigint NOT NULL DEFAULT 0;