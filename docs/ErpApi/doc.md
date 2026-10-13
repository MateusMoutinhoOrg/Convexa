# ErpApi

Every screen of Convexa has a JSON twin, so anything a person does on a page a program can do through the API.

## Authentication

1. Sign in, open **API tokens**, and create a token.
2. Send it on every call:

   ```
   Authorization: Bearer bo_…
   ```

A token acts with its user's current role:

| Prefix | Who can call it |
|---|---|
| `POST /api/admin/erp/<call>` | any token; these calls only read |
| `POST /api/admin/manage/<call>` | admin and root tokens; these calls write |
| `/api/admin/root/<call>` | root tokens only; these are users, backups and settings |

## Conventions

- Every call is a `POST` with a JSON body, and answers JSON.
- Keys are kebab-case and in English, whatever the language: `third-party`, `due-date`.
- Amounts are integer **cents**. Dates are `2026-10-12` and months are `2026-10`. An empty string means no date.
- A record names other records by id, and `0` means none.
- `add-*` answers `201`. `set-*` rewrites **every** field, so send the whole record, as `get-*` answers it. `remove-*` answers `{"status":"ok"}`.
- A refusal answers `{"error": "…", "field": "…"}`:
  - `400`: invalid input;
  - `403`: the role does not allow it;
  - `404`: not found;
  - `409`: the record is in use, or there is nothing left to pay.
- Messages follow `Accept-Language` or `?lang=pt`, so `Accept-Language: pt-BR` answers in Portuguese.

## Calls

| Area | Read (`/api/admin/erp/`) | Write (`/api/admin/manage/`) |
|---|---|---|
| Companies | `list-companies`, `get-company` | `add-company`, `set-company`, `remove-company` |
| Accounts | `list-accounts` (with balances), `get-account` | `add-account`, `set-account`, `remove-account` |
| Categories | `list-categories`, `get-category` | `add-category`, `set-category`, `remove-category` |
| Third parties | `list-third-parties`, `get-third-party` | `add-third-party`, `set-third-party`, `remove-third-party` |
| Transactions | `list-transactions`, `get-transaction` | `add-transaction`, `set-transaction`, `remove-transaction`, `settle-transaction`, `unsettle-transaction`, `add-transfer`, `set-transfer` |
| Credit cards | `list-credit-cards`, `list-card-invoices`, `get-card-invoice` | `pay-card-invoice` |
| Recurrences | `list-recurrences`, `get-recurrence` | `add-recurrence`, `set-recurrence`, `remove-recurrence`, `generate-recurrences` |
| Budgets | `list-budgets`, `get-budget` | `add-budget`, `set-budget`, `remove-budget` |
| Reports | `dashboard`, `payables-receivables`, `cash-flow`, `projection`, `company-report`, `account-balances`, `account-statement`, `budget-report` | |
| Settings | `get-settings` | `/api/admin/root/set-settings` (root) |

[Routes](../Routes/doc.md) and `GET /openapi.json` list every body and every answer.

## Examples

```bash
H='-H "Authorization: Bearer $TOKEN" -H Content-Type:application/json'

# A company, an account with its opening balance, a card
curl -X POST localhost:3000/api/admin/manage/add-company $H -d '{"name":"Personal"}'
curl -X POST localhost:3000/api/admin/manage/add-account $H -d '{"name":"Checking","kind":"checking","opening-balance":150000,"opening-company":1,"opening-date":"2026-10-01"}'
curl -X POST localhost:3000/api/admin/manage/add-account $H -d '{"name":"Visa","kind":"credit-card","credit-limit":500000,"closing-day":3,"due-day":10,"payment-account":1}'

# A bill to pay, then paid
curl -X POST localhost:3000/api/admin/manage/add-transaction $H -d '{"company":1,"account":1,"type":"expense","amount":25000,"description":"Electricity","date":"2026-10-12","due-date":"2026-10-20","status":"pending"}'
curl -X POST localhost:3000/api/admin/manage/settle-transaction $H -d '{"id":4,"settled-date":"2026-10-19"}'

# A purchase in 10 installments on the card, and its invoice paid
curl -X POST localhost:3000/api/admin/manage/add-transaction $H -d '{"company":1,"account":2,"type":"expense","amount":120000,"description":"TV","date":"2026-10-12","status":"settled","installments":10}'
curl -X POST localhost:3000/api/admin/manage/pay-card-invoice $H -d '{"id":2,"month":"2026-11","date":"2026-11-10","status":"settled"}'

# The projected balance, a year ahead, with every event to come
curl -X POST localhost:3000/api/admin/erp/projection $H -d '{"months":12,"events":true}'
```
