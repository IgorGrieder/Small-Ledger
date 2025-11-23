-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- 1. Create the Accounts
INSERT INTO accounts (id, name, currency, metadata) VALUES 
-- Igor has a USD Account
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Igor', 'USD', '{"type": "user"}'),

-- Maria Clara has a BRL Account
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Maria Clara', 'BRL', '{"type": "user"}');


-- 2. Create "Deposit" Transaction Header
INSERT INTO transactions (id, description, status) VALUES 
('33333333-3333-3333-3333-333333333333', 'Initial Deposit for Igor and Maria', 'posted');


-- 3. Create Entries (World Bank -> Users)
-- remember: World Bank is 000...000 (created in previous step)
INSERT INTO entries (id, transaction_id, account_id, amount, currency) VALUES 

-- Give Igor $1,000.00 USD
(gen_random_uuid(), '33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000000', -100000, 'USD'), -- Debit World
(gen_random_uuid(), '33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',  100000, 'USD'), -- Credit Igor

-- Give Maria Clara R$5,000.00 BRL
(gen_random_uuid(), '33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000009', -500000, 'BRL'), -- Debit World
(gen_random_uuid(), '33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',  500000, 'BRL'); -- Credit Maria
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
