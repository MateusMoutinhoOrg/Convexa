# Erp

This page explains how Convexa models money. The code lives in `sandbox/internal/erp/`:

| Package | Holds |
|---|---|
| `erpstore` | writes and validation |
| `erpledger` | arithmetic |
| `erpreports` | reports |
| `erpview` | pages |
| `erpapi` | JSON |
| `erpi18n` | languages |
| `erpdate` | dates |
| `erpmoney` | money |
| `erpjobs` | the background job |

The records live in the `erp-db` database. See `sandbox/internal/databases/erp_db/database.yaml`.

## Companies

Everything belongs to a company. A person who keeps their own finances registers themselves as one, such as "Personal". Nothing tells a person from a business.

A company cannot be removed while a transaction, recurrence or budget names it. **Archive** it instead: an archived record keeps its history but is not offered for new entries. Accounts, categories and third parties follow the same rule.

## Money and dates

- Amounts are **integer cents**, so `123456` is 1,234.56. They are never floats.
- A transaction's amount is **signed**: an inflow is positive, an outflow negative. The API takes a positive `amount` for income, expense and transfers, and the type gives the sign. An adjustment takes the sign you send.
- Dates are stored as `YYYYMMDD` integers and spelled `2026-10-12` in the API. Months are spelled `2026-10`.
- "Today" is read `--utc-offset` minutes away from UTC.

## Accounts

An account belongs to no company. Several companies can move money through one account, and each transaction says whose money it moves. So an account has two kinds of balance:
- **absolute**: the cash it holds, every company counted;
- **per company**: the same sum over that company's transactions only.

The absolute balance is always the sum of the per-company balances.

A balance at a date counts what was **settled** by then.

Kinds: `checking`, `savings`, `investment`, `wallet`, `credit-card`, `other`. **Cash** means every account except the credit cards.

An opening balance is an `adjustment` transaction. The account form offers one for the first company. Add the other companies' shares as more adjustments.

## Transactions

| Field | Meaning |
|---|---|
| `type` | `income`, `expense`, `adjustment`, `transfer-in`, `transfer-out` |
| `date` | the date it is entered for (competence) |
| `due-date` | when it is due; blank means `date` |
| `status` | `pending` (a payable or receivable) or `settled` |
| `settled-date` | when it was paid or received |

A pending transaction past its due date is **overdue**.

**Installments:** `installments: N` splits the amount into N transactions, one month apart. Any leftover cent goes on the first. All N share an `installment-group`, which is the id of the first.

**Transfers** are two transactions pointing at each other through `counterpart`: a `transfer-out` leg and a `transfer-in` leg.
- A transfer may move money between two accounts, or between two companies on the same account.
- Its two legs always share the same dates and status.
- Neither leg counts as income or expense.

## Credit cards

A card has a `credit-limit`, a `closing-day`, a `due-day` and a `payment-account`.

**Invoices.** Invoice *M* closes on the closing day of month *M*. It is due on the due day of *M*, or of the next month when the due day comes before the closing day. Days are clamped to the month's length.

**Purchases.** A purchase is an `expense` on the card, and it is settled the day it is made, because the card owes it from then on. It lands on the invoice of the month it was made in, up to the closing day, and on the next invoice after that. Installment *k* of a purchase lands on invoice *M + k*.

**Paying an invoice** is a transfer from the payment account to the card, tagged with the invoice's month. *Pay invoice* writes one transfer per company, each paying that company's share, so every company's balance on the card returns to zero.

**Invoice status** is derived from the invoice's charges and payments: `open`, `closed`, `partial`, `paid` or `overdue`.

## Recurrences

A recurrence repeats every `interval` weeks, months or years from its `start-date`, until its `end-date` if it has one. A monthly recurrence on the 31st falls on the last day of shorter months.

**Writing occurrences ahead.** Occurrences up to `--recurrence-horizon` days ahead are written as pending transactions, so they show up as payables and receivables before they are due. This happens:
- when the server starts;
- every 6 hours after that;
- on *Generate now*.

`generated-until` is the last day already written. Occurrences beyond it count only in the projection.

**Editing a recurrence** rewrites its pending occurrences from today on. Settled and overdue occurrences stay as they are.

**Removing a recurrence** deletes its pending occurrences from today on. The transactions it wrote before stay, but no longer point at it.

## Budgets

A budget is the amount a company plans to spend, or to receive, in one category and its subcategories over a period. *Repeat monthly* writes one budget per month.

What a budget has **committed** is the sum, as a positive amount, of the transactions in its period:
- of its company;
- in its category tree;
- dated in its period, by `date`;
- settled or pending;
- on any account, cards included.

## Reports

- **Events.** Every report is a sum of *events*: settled transactions, pending ones, recurrence occurrences not written yet, what is left of card invoices, and what is left of budgets. A pending event that is overdue is placed on today.
- **Cash flow** covers the cash accounts, by day, week or month.
  - The *realized* basis counts settled transactions only.
  - The *projected* basis adds everything still expected.
  - Transfers between accounts in view cancel out.
  - Card purchases enter when their invoice is paid.
- **Projected balance** starts from today's settled cash. It adds, day by day, every pending transaction, every recurrence occurrence, and whatever is left of each card invoice, on the account that pays the card.
  - It is drawn twice, *with* and *without* budgets.
  - "With budgets" adds what is left of each budget on the last day of its period: the budget less what it has committed, and nothing once a budget is reached.
- **Payables & receivables** lists pending transactions and the card invoices still to pay, one per company.
- **By company** shows income and expense by category, balances, and budgets.
- **By account** shows the matrix of absolute and per-company balances, plus statements.

## Roles

| Role | Pages and API it can reach |
|---|---|
| root | everything, including `/admin/root/*` and `/api/admin/root/*`: users, backups, settings |
| admin | everything under `/admin/manage/*` and `/api/admin/manage/*`: every financial record |
| basic | `/admin/erp/*` and `/api/admin/erp/*`: reads everything, changes nothing |

The guards are the `erp-manage-guard` and `erp-api-manage-guard` middlewares, beside the backoffice's root guards.

## Consistency

The store has no multi-record transactions.
- Every write goes through `erpstore`, which validates the input and serializes writes behind one lock.
- A write of several records, such as transfer legs, installments or invoice payments, removes what it wrote when a step fails.
- The transaction listing is cached in memory. Every write invalidates the cache, and so does a backup restore.
