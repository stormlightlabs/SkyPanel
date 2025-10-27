CREATE TABLE IF NOT EXISTS posts (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    uri TEXT NOT NULL UNIQUE,
    author_did TEXT NOT NULL,
    text TEXT NOT NULL,
    feed_id TEXT NOT NULL,
    indexed_at DATETIME NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_feed_id ON posts(feed_id);
CREATE INDEX IF NOT EXISTS idx_posts_author_did ON posts(author_did);
CREATE INDEX IF NOT EXISTS idx_posts_indexed_at ON posts(indexed_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_uri ON posts(uri);
