# Ledger Service (Multi-Currency) -- Project Overview

## How to run

First, run the migration to create the PostgreSQL structure with goose!

```
goose -dir ./internal/db/migrations postgres "user=postgres password=none dbname=ledger host=localhost port=5432 sslmode=disable" up
```

## 📌 Introduction

This project implements a **minimal, reliable, multi-currency ledger
service** designed to track monetary movements across accounts.\
It is intentionally simple, but structured following real accounting
principles (double-entry), ensuring **auditability**, **correctness**,
and **immutability**.

The ledger supports: - **USD (US Dollar)** - **BRL (Brazilian Real)**\
...but the design is general enough to add more currencies later.

---

## 🎯 Goals

The goal of this project is to provide a foundational ledger system
that:

1. Records all money movements in an **append-only**, **auditable**
   format.
2. Supports **double-entry accounting**: every transaction affects two
   or more accounts.
3. Correctly manages **multi-currency entries**.
4. Calculates account balances from the ledger itself (source of
   truth).
5. Provides clean APIs or functions to post transactions and query
   balances.
6. Ensures that money movements are **atomic**, **consistent**, and
   **idempotent**.

---

## 🧱 Core Concepts

### 1. **Accounts**

An account represents a place where money lives.\
Examples: - User wallet - System reserve - FX pool - Merchant settlement
account

Each account has: - `id` - `name` - `currency` (USD or BRL) - optional
metadata

Accounts do **not** store balances directly---balances are derived from
entries.

---

### 2. **Transactions**

A **transaction** is a _logical event_ that groups money movements.

Examples: - Transfer from User A → User B - Deposit or withdrawal - FX
conversion (USD ↔ BRL) - Refund / reversal

A transaction includes: - `id` - `description` - `external_id` (for
idempotency) - `status` - `created_at` - metadata

A transaction contains **multiple entries**.

---

tries\*\*

Entries are the heart of the ledger --- each row represents a single
movement of money.

Each entry contains: - `transaction_id` - `account_id` - `amount`
(positive or negative) - `currency` - timestamp + metadata

**A single transaction must have entries that sum to zero _per
currency_.**

Examples:

#### BRL Transfer Example

    transaction: tx_001

    entries:
    - Account B: +100 BRL

#### FX Example (USD → BRL)

    transaction: tx_fx_01

    entries:
    - Debit user USD account
    - Credit FX-USD pool
    - Debit FX-BRL pool
    - Credit user BRL account
    - Fee entries (optional)

---

## 🔒 Invariants

- Ledger is append-only (no deletes, no updates to entries).
- Each transaction must be **balanced**:
  - Sum(entries.amount where currency = USD) = 0\
  - Sum(entries.amount where currency = BRL) = 0
- Posting a transaction is atomic.
- Account balances must always match the ledger sum.

---

## 📊 Balance Calculation

Balance = `SUM(entries.amount)` for that account and currency.

For performance, optional optimizations: - Materialized balance table
updated in the same DB transaction. - Snapshots every N entries.

---

## 🧪 Error Handling & Idempotency

The system SHOULD: - Reject unbalanced transactions - Reject entries
posted to accounts of mismatched currency - Be idempotent via
`external_id`

Example: If the client retries a failed HTTP call, the ledger returns
the same transaction instead of duplicating entries.

---

## 🔁 Reversals

Instead of "undoing" or deleting anything, the system posts a **new
transaction** with the inverted entries.

Example reversal of a BRL transfer:

    original:  A -100, B +100
    reversal:  A +100, B -100

---

## 🏗 Future Extensions

- Webhooks on new transactions
- FX rate engine
- Scheduled transactions
- CSV/Excel export
- Partitioned tables for high volume
- Cryptographic ledger hashes (blockchain-style immutability)

---

## ✔ Summary

This project implements a robust yet minimal ledger system based on: -
**Accounts** - **Transactions** - **Entries**

It ensures correct, auditable, multi-currency financial operations and
forms the foundation for wallets, banking layers, payment systems, and
financial infrastructure.
