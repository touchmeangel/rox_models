SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

CREATE TABLE IF NOT EXISTS users (
    id            text        PRIMARY KEY,
    email         text        NOT NULL,
    username      text        NOT NULL,
    roles         text[]      NOT NULL DEFAULT '{}',
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (lower(email));
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (lower(username));

CREATE TABLE IF NOT EXISTS runs (
    id                     text        PRIMARY KEY,
    name                   text        NOT NULL,
    user_id                text        NOT NULL REFERENCES users(id),
    status                 text        NOT NULL,
    workspace_name         text        NOT NULL DEFAULT '',
    worker_count           bigint      NOT NULL DEFAULT 0,
    completed_worker_count bigint      NOT NULL DEFAULT 0,
    active_worker_count    bigint      NOT NULL DEFAULT 0,
    created_at             timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_runs_user_id_created_at ON runs (user_id, created_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_user_id_name ON runs (user_id, name);

CREATE TABLE IF NOT EXISTS coordinators (
    id         text        PRIMARY KEY,
    run_id     text        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    active     boolean     NOT NULL DEFAULT false,
    completed  boolean     NOT NULL DEFAULT false,
    error      text        NOT NULL DEFAULT '',
    retriable  boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_coordinators_run_id ON coordinators (run_id);

CREATE TABLE IF NOT EXISTS workers (
    id         text        PRIMARY KEY,
    run_id     text        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    mission_id text        NOT NULL,
    active     boolean     NOT NULL DEFAULT false,
    completed  boolean     NOT NULL DEFAULT false,
    error      text        NOT NULL DEFAULT '',
    retriable  boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workers_run_id ON workers (run_id);