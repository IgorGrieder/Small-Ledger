-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- 1. Create Accounts
INSERT INTO accounts (id, name, currency, metadata) VALUES 
-- The "World" (Infinite Source)
('00000000-0000-0000-0000-000000000000', 'World Bank', 'USD', '{"type": "equity"}'),
('00000000-0000-0000-0000-000000000009', 'World Bank', 'BRL', '{"type": "equity"}');

-- 2. Create Transaction Header
INSERT INTO transactions (id, description, status) VALUES 
('11111111-1111-1111-1111-111111111111', 'Seed Funding for Pools', 'posted');

-- 3. Create Entries (Move Money World -> Pools)
INSERT INTO entries (id, transaction_id, account_id, amount, currency) VALUES
-- Fund USD Pool
(gen_random_uuid(), '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000000', -100000000, 'USD'), -- Debit World
(gen_random_uuid(), '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001',  100000000, 'USD'), -- Credit Pool

-- Fund BRL Pool
(gen_random_uuid(), '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000009', -100000000, 'BRL'), -- Debit World
(gen_random_uuid(), '11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000002',  100000000, 'BRL'); -- Credit Pool
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
