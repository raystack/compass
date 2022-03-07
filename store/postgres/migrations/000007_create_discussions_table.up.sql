CREATE TABLE discussions (
  id serial PRIMARY KEY,
  title text NOT NULL,
  body text NOT NULL,
  state text NOT NULL,
  type text NOT NULL,
  owner uuid NOT NULL,
  labels text[],
  assets text[],
  assignees text[],
  created_at timestamp DEFAULT NOW(),
  updated_at timestamp DEFAULT NOW()
);

CREATE INDEX discussions_idx_assets ON discussions USING GIN(assets);
CREATE INDEX discussions_idx_labels ON discussions USING GIN(labels);
CREATE INDEX discussions_idx_assignees ON discussions USING GIN(assignees);
CREATE INDEX discussions_idx_owner ON discussions(owner);
CREATE INDEX discussions_idx_type_state ON discussions(type,state);
