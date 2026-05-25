ALTER TABLE knowledge_items
  ADD COLUMN IF NOT EXISTS access_count bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_accessed_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_knowledge_items_access_count
ON knowledge_items(access_count DESC);
