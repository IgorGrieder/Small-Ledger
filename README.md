# Ledger vs Account Balance — Explanation & Diagram

## Short answer

- **The ledger is the source of truth.** The canonical balance for any account is the **sum of the ledger entries** for that account (credits increase the account, debits decrease it, depending on your convention).
- A stored `accounts.balance` column can be kept as a **materialized/derived** optimization for fast reads, but it is not the authoritative record — the ledger entries are.

---

## Why ledger-first?

1. **Auditability:** ledger entries are immutable historical records. You can show an auditor the sequence of entries that produced a balance.
2. **Correctness:** double-entry accounting (every transaction has equal debits and credits) preserves system-wide invariants (total money conserved unless external settlement occurs).
3. **Reversals & fixes:** instead of editing history, you post compensating transactions that are visible and auditable.

---

## How balance is calculated (formula)

Use amounts in minor units (cents) and keep ledger entries immutable.

**Query to compute authoritative balance from the ledger**

```sql
SELECT
  account_id,
  SUM(CASE WHEN side = 'credit' THEN amount ELSE -amount END) AS balance_cents
FROM ledger_entries
WHERE account_id = $1
GROUP BY account_id;
```

If your `ledger_entries` encode sign differently (e.g., positive for both debits and credits but with a `side`), adapt the CASE accordingly.

**Alternative (using `amount` with sign):**
If you store signed `amount` (positive for credit, negative for debit):

```sql
SELECT SUM(amount) AS balance_cents FROM ledger_entries WHERE account_id = $1;
```

---

## When to keep `accounts.balance` (materialized)

For small projects, you can keep `accounts.balance` to serve fast reads and APIs, but follow these rules:

- Update `accounts.balance` **within the same DB transaction** that writes the ledger entries. That ensures the materialized balance and entries move together atomically.
- Periodically (nightly or after a batch of transactions) **reconcile** the materialized balances against the authoritative ledger (`SUM(entries)`), and alert if there's drift.
- Prefer storing `balance` in minor units (`BIGINT`) and avoid floats.

**Advantages:** faster reads, simple APIs.

**Disadvantages:** risk of drift if updates fail or bugs exist; more complexity to keep reconciliation tooling.

---

## Recommended patterns (practical)

- Use `SELECT ... FOR UPDATE` on affected account rows inside a DB transaction when posting transactions. This prevents concurrent races when updating `accounts.balance`.
- Keep ledger entries immutable — never update or delete. Reversals are new transactions.
- Use an `idempotency_keys` table to map client idempotency keys to transaction ids so retries don't duplicate ledger entries.
- Keep an `outbox` table in the same transaction if you need to publish events — this avoids distributed transaction issues.

---

## Simple example (Alice → Bob, step-by-step)

1. Alice (A) balance before: 5000 cents (R$50). Bob (B) balance before: 2000 cents (R$20).
2. Post transfer of 1000 cents from A to B.
3. In single DB tx: insert `transactions`, insert two `ledger_entries` (debit A 1000, credit B 1000), update `accounts.balance` for A and B accordingly.
4. After commit, authoritative balances (sum of ledger entries) equal the stored `accounts.balance`.

---

## Reconciliation check (simple SQL)

```sql
-- find accounts where materialized balance differs from sum(entries)
SELECT a.id,
       a.balance AS materialized_balance,
       COALESCE(l.sum_entries, 0) AS ledger_sum
FROM accounts a
LEFT JOIN (
  SELECT account_id, SUM(CASE WHEN side='credit' THEN amount ELSE -amount END) AS sum_entries
  FROM ledger_entries
  GROUP BY account_id
) l ON l.account_id = a.id
WHERE a.balance <> COALESCE(l.sum_entries, 0);
```

If this query returns rows, you have drift — investigate cause (failed updates, out-of-sync code paths, bugs).

---

## The diagram (draw.io XML)

Below is a small draw.io (diagrams.net) XML. To import it:

1. Copy everything inside the code fence into a file named `ledger-diagram.drawio` (or `ledger-diagram.xml`).
2. Open [https://app.diagrams.net/](https://app.diagrams.net/)
3. File → Import From → Device → select your `ledger-diagram.drawio` file.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<mxfile host="app.diagrams.net" modified="2025-10-25T00:00:00.000Z" agent="" etag="" version="20.0.0" type="device">
  <diagram id="ledger" name="Ledger Flow">
    <mxGraphModel dx="1280" dy="720" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>

        <!-- Client -->
        <mxCell id="client" value="Client / UI / Service" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#1a73e8;" vertex="1" parent="1">
          <mxGeometry x="50" y="60" width="160" height="60" as="geometry"/>
        </mxCell>

        <!-- API Gateway -->
        <mxCell id="api" value="API Gateway / Auth" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#34a853;" vertex="1" parent="1">
          <mxGeometry x="260" y="60" width="160" height="60" as="geometry"/>
        </mxCell>

        <!-- Service -->
        <mxCell id="service" value="Application Service\n(Validates, Idempotency)" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#fbbc05;" vertex="1" parent="1">
          <mxGeometry x="470" y="60" width="200" height="60" as="geometry"/>
        </mxCell>

        <!-- Ledger Core -->
        <mxCell id="ledgercore" value="Ledger Core / Accounting Engine\n(creates entries, enforces balance)" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#ea4335;" vertex="1" parent="1">
          <mxGeometry x="740" y="30" width="260" height="100" as="geometry"/>
        </mxCell>

        <!-- DB -->
        <mxCell id="db" value="Postgres\n(tables: accounts, transactions, ledger_entries, idempotency_keys, outbox)" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#5f6368;" vertex="1" parent="1">
          <mxGeometry x="1050" y="30" width="260" height="120" as="geometry"/>
        </mxCell>

        <!-- Outbox -->
        <mxCell id="outbox" value="Outbox → Event bus" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#9c27b0;" vertex="1" parent="1">
          <mxGeometry x="1050" y="170" width="160" height="60" as="geometry"/>
        </mxCell>

        <!-- Admin / Auditor -->
        <mxCell id="admin" value="Admin / Auditor\n(Export / Reconcile)" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;" vertex="1" parent="1">
          <mxGeometry x="50" y="200" width="160" height="60" as="geometry"/>
        </mxCell>

        <!-- Edges -->
        <mxCell id="e1" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="client" target="api">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="e2" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="api" target="service">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="e3" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="service" target="ledgercore">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="e4" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="ledgercore" target="db">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="e5" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="db" target="outbox">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="e6" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;" edge="1" parent="1" source="db" target="admin">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
```

---

## Final tips about the account `balance` field

- Treat `accounts.balance` as a **cache/optimization**. If you ever suspect it might be stale, re-calculate from the ledger. Keep reconciliation tooling accessible via an admin endpoint or a scheduled job.
- Always update `accounts.balance` inside the same DB transaction that writes `ledger_entries`. If you do that, `accounts.balance` will be consistent after commit.
- Keep small handy scripts (or SQL) that can recompute and _fix_ drift for the ops team — but do this with care and audit every manual correction.

---

If you want, I can also:

- Produce a PNG/SVG export of the diagram you can download directly.
- Generate the minimal SQL and Node.js handler (or a small repo) wired to this architecture.
