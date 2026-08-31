SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE runs DROP COLUMN IF EXISTS worker_count;
ALTER TABLE runs DROP COLUMN IF EXISTS completed_worker_count;
ALTER TABLE runs DROP COLUMN IF EXISTS active_worker_count;