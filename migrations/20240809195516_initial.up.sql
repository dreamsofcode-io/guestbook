CREATE TABLE guest (
  id uuid primary key,
  message varchar(256) not null,
  ip inet not null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

CREATE INDEX ON guest (created_at);
