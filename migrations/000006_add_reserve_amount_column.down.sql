SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE upload_sessions DROP COLUMN IF EXISTS reserve_amount;