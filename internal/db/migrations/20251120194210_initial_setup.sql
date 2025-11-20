-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TYPE currency AS ENUM ('USD', 'BRL');

-- 1. ACCOUNTS
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    currency currency NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. TRANSACTIONS
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    external_id TEXT UNIQUE,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'posted',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. ENTRIES
CREATE TABLE entries (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount NUMERIC(20,4) NOT NULL,
    currency currency NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. VALIDATION FUNCTION
CREATE OR REPLACE FUNCTION validate_entry_currency()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT currency FROM accounts WHERE id = NEW.account_id) != NEW.currency THEN
        RAISE EXCEPTION 'Entry currency (%) must match Account currency', NEW.currency;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 5. TRIGGER
CREATE TRIGGER check_currency_match
BEFORE INSERT OR UPDATE ON entries
FOR EACH ROW
EXECUTE FUNCTION validate_entry_currency();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS check_currency_match ON entries;
DROP FUNCTION IF EXISTS validate_entry_currency();
DROP TABLE entries;
DROP TABLE transactions;
DROP TABLE accounts;
DROP TYPE currency;
-- +goose StatementEnd
