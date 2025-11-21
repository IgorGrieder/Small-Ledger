-- name: GetAllTransactions :many
SELECT * from transactions;

-- name: GetAllAccounts :many
SELECT * from accounts;

-- name: GetAllEntries :many
SELECT * from entries;

-- name: GetUserFunds :one
SELECT SUM(entries.amount) as Funds from entries where account_id = $1;
