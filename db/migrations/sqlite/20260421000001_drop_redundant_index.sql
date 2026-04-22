-- Drop redundant index on links(key)
-- The PRIMARY KEY already creates an index, so idx_links_key is unnecessary

DROP INDEX IF EXISTS idx_links_key;
