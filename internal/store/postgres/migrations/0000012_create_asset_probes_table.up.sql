CREATE TABLE asset_probes (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  asset_urn text NOT NULL REFERENCES assets(urn) ON DELETE CASCADE ON UPDATE CASCADE,
  status text NOT NULL,
  status_reason text,
  metadata jsonb,
  timestamp timestamp DEFAULT NOW(),
  created_at timestamp DEFAULT NOW()
);
