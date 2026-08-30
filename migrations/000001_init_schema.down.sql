SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

DROP TABLE IF EXISTS workers;
DROP TABLE IF EXISTS coordinators;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS users;