CREATE TABLE IF NOT EXISTS users
(
    id INTEGER PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    user_id TEXT NOT NULL UNIQUE,
    pass_hash BLOB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login ON users (login);

CREATE TABLE IF NOT EXISTS apps
(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

-- Insert a new record into the apps table after creating it

INSERT INTO apps (id, name, secret)
VALUES (1,'app', 'secret')
ON CONFLICT DO NOTHING;