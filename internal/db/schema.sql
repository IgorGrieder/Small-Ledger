CREATE TYPE currency AS ENUM ('USD', 'BRL');

-- =========================================
-- ACCOUNTS
-- =========================================
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    currency currency NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =========================================
-- TRANSACTIONS (logical grouping)
-- =========================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    external_id TEXT UNIQUE,                       -- idempotency key
    description TEXT,
    status TEXT NOT NULL DEFAULT 'posted',         -- posted | pending | reversed
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =========================================
-- ENTRIES (individual money movements)
-- =========================================
CREATE TABLE entries (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount NUMERIC(20,4) NOT NULL,                 -- positive or negative
    currency currency NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
