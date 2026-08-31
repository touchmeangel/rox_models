SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE system_settings (
    admin_claimed       boolean NOT NULL DEFAULT false,
    open_signup_enabled boolean NOT NULL DEFAULT false
);

INSERT INTO system_settings (admin_claimed, open_signup_enabled) VALUES (false, false);