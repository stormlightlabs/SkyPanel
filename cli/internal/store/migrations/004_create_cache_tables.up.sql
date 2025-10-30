-- Follower snapshots metadata
CREATE TABLE IF NOT EXISTS follower_snapshots (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    user_did TEXT NOT NULL,
    snapshot_type TEXT NOT NULL,
    total_count INTEGER NOT NULL,
    expires_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snapshots_user_type ON follower_snapshots(user_did, snapshot_type);
CREATE INDEX IF NOT EXISTS idx_snapshots_created ON follower_snapshots(created_at);
CREATE INDEX IF NOT EXISTS idx_snapshots_expires ON follower_snapshots(expires_at);

-- Snapshot entries (actors in each snapshot)
CREATE TABLE IF NOT EXISTS follower_snapshot_entries (
    snapshot_id TEXT NOT NULL,
    actor_did TEXT NOT NULL,
    indexed_at TEXT,
    PRIMARY KEY(snapshot_id, actor_did),
    FOREIGN KEY(snapshot_id) REFERENCES follower_snapshots(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snapshot_entries_actor ON follower_snapshot_entries(actor_did);

-- Cached post rate metrics
CREATE TABLE IF NOT EXISTS cached_post_rates (
    actor_did TEXT PRIMARY KEY,
    posts_per_day REAL NOT NULL,
    last_post_date DATETIME,
    sample_size INTEGER NOT NULL,
    fetched_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_post_rates_fetched ON cached_post_rates(fetched_at);
CREATE INDEX IF NOT EXISTS idx_post_rates_expires ON cached_post_rates(expires_at);

-- Cached activity data (last post dates)
CREATE TABLE IF NOT EXISTS cached_activity (
    actor_did TEXT PRIMARY KEY,
    last_post_date DATETIME,
    fetched_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_activity_fetched ON cached_activity(fetched_at);
CREATE INDEX IF NOT EXISTS idx_activity_expires ON cached_activity(expires_at);
