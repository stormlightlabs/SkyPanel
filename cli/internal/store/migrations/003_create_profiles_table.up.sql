CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    did TEXT NOT NULL UNIQUE,
    handle TEXT NOT NULL,
    data_json TEXT NOT NULL,
    fetched_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_profiles_did ON profiles(did);
CREATE INDEX IF NOT EXISTS idx_profiles_handle ON profiles(handle);
CREATE INDEX IF NOT EXISTS idx_profiles_fetched_at ON profiles(fetched_at);
