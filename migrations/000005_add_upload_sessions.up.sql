SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE upload_sessions (
    run_id          text PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
    user_id         text NOT NULL,
    upload_id       text NOT NULL,
    offset_bytes    bigint NOT NULL DEFAULT 0,
    completed_parts jsonb NOT NULL DEFAULT '[]',
    last_active_at  timestamptz NOT NULL DEFAULT now()
);