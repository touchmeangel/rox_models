SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE users DROP COLUMN IF EXISTS registered_at;
ALTER TABLE users DROP COLUMN IF EXISTS used_bytes;
ALTER TABLE users DROP COLUMN IF EXISTS quota_bytes;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;