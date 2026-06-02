CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS admin_users (
  id bigserial PRIMARY KEY,
  username varchar(64) UNIQUE NOT NULL,
  password_hash varchar(255) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  username varchar(64) UNIQUE NOT NULL,
  password_hash varchar(255) NOT NULL,
  nickname varchar(100),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_items (
  id bigserial PRIMARY KEY,
  title varchar(200) NOT NULL,
  question text NOT NULL,
  answer text NOT NULL,
  category varchar(100),
  tags text[],
  source_type varchar(32) NOT NULL DEFAULT 'faq',
  status varchar(32) NOT NULL DEFAULT 'approved',
  access_count bigint NOT NULL DEFAULT 0,
  last_accessed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_chunks (
  id bigserial PRIMARY KEY,
  item_id bigint NOT NULL REFERENCES knowledge_items(id) ON DELETE CASCADE,
  chunk_text text NOT NULL,
  embedding vector(1024),
  source_url text,
  page_no integer,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_submissions (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  question text NOT NULL,
  answer text NOT NULL,
  category varchar(100),
  tags text[],
  source text,
  remark text,
  status varchar(32) NOT NULL DEFAULT 'pending',
  reviewer_note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  reviewed_at timestamptz
);

CREATE TABLE IF NOT EXISTS query_logs (
  id bigserial PRIMARY KEY,
  user_id bigint REFERENCES users(id) ON DELETE SET NULL,
  user_question text NOT NULL,
  normalized_question text,
  intent varchar(64),
  matched_item_id bigint REFERENCES knowledge_items(id) ON DELETE SET NULL,
  hit_score numeric(5,4),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_knowledge_items_status ON knowledge_items(status);
CREATE INDEX IF NOT EXISTS idx_knowledge_items_category ON knowledge_items(category);
CREATE INDEX IF NOT EXISTS idx_knowledge_items_source_type ON knowledge_items(source_type);
CREATE INDEX IF NOT EXISTS idx_knowledge_items_created_at ON knowledge_items(created_at);
CREATE INDEX IF NOT EXISTS idx_knowledge_items_access_count ON knowledge_items(access_count DESC);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_item_id ON knowledge_chunks(item_id);

CREATE INDEX IF NOT EXISTS idx_user_submissions_user_id ON user_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_submissions_status ON user_submissions(status);
CREATE INDEX IF NOT EXISTS idx_user_submissions_created_at ON user_submissions(created_at);

CREATE INDEX IF NOT EXISTS idx_query_logs_user_id ON query_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_query_logs_created_at ON query_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_query_logs_matched_item_id ON query_logs(matched_item_id);
CREATE INDEX IF NOT EXISTS idx_query_logs_intent ON query_logs(intent);

-- Create this after enough rows exist in knowledge_chunks for ivfflat training.
-- CREATE INDEX idx_knowledge_chunks_embedding
-- ON knowledge_chunks
-- USING ivfflat (embedding vector_cosine_ops)
-- WITH (lists = 100);
