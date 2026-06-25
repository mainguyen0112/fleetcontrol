CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE satellites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    region TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Pending',
    managed_by TEXT NOT NULL DEFAULT 'manual',
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);