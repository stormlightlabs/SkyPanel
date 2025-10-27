CREATE TABLE IF NOT EXISTS feeds (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    params TEXT,
    is_local BOOLEAN NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_feeds_name ON feeds(name);
CREATE INDEX IF NOT EXISTS idx_feeds_source ON feeds(source);
