-- Initial schema for golinks
-- Creates the links table and schema_migrations tracking table

CREATE TABLE IF NOT EXISTS links (
    key TEXT PRIMARY KEY,
    target TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_links_key ON links(key);
