CREATE TABLE comments (
  id serial PRIMARY KEY,
  discussion_id serial NOT NULL REFERENCES discussions(id) ON DELETE CASCADE,
  body text NOT NULL,
  owner uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  updated_by uuid NOT NULL,
  created_at timestamp DEFAULT NOW(),
  updated_at timestamp DEFAULT NOW()
);

CREATE INDEX comments_idx_discussion_id ON comments(discussion_id);
CREATE INDEX comments_idx_owner ON comments(owner);
