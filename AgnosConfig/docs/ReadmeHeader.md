# Convexa

[![Go Reference](https://pkg.go.dev/badge/{{.Module}}.svg)](https://pkg.go.dev/{{.Module}})
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](go.mod)

Convexa is a financial ERP for companies, and for people too: a person who keeps their own finances registers themselves as a company ("Personal"). It is built with [agnos](https://github.com/MateusMoutinhoOrg/Agnos) in backoffice mode and ships as one binary. That binary serves the web interface and a JSON API, and keeps its data in plain files under a `data` folder.

Everything in Convexa revolves around **companies**:
- **Accounts**: checking, savings, investment, wallet and credit card. One account can be shared by several companies. Its balance is shown both as a whole (the cash it holds) and as each company's share.
- **Third parties**: customers, suppliers, or both.
- **Categories** classify income and expense, and can be nested.
- **Transactions** (lançamentos): income, expense, transfers and adjustments. Each one has an entry date, a due date and a settlement date. A pending transaction is a payable or a receivable.
- **Credit cards** have a limit, a closing day and a due day. Purchases, including purchases in installments, land on monthly invoices. Paying an invoice is a transfer, split by company.
- **Recurrences**: rent, salaries and subscriptions are written as pending transactions ahead of time.
- **Budgets** are set per company, category and period, or month by month.
- **Reports**: dashboard, cash flow, payables and receivables, company report, account balances and statements, budget vs actual, and the **projected balance** up to 5 years ahead, with and without budgets.
- **Roles**:
  - **root** controls everything, including users and backups.
  - **admin** manages every financial record.
  - **basic** only reads.
- **Languages**: everything is written in English. Every page and every API message can be switched to Portuguese.

## Quick start

```bash
go build -o convexa ./cmd/main                       # or: agnos build
export CONVEXA_BACKOFFICE_SECRET=$(openssl rand -hex 32)
./convexa add-backoffice-user --username root --email you@example.com --role root   # prints the password once
./convexa start-server --insecure-http --addr 127.0.0.1:3000 --utc-offset -180
```

Open <http://127.0.0.1:3000/admin>, sign in, and start by adding a company and an account.

Flags of `start-server`:

| Flag | What it does |
|---|---|
| `--insecure-http` | Serves plain http, for local use. Without it, run Convexa behind a TLS proxy. |
| `--utc-offset` | Minutes from UTC that "today" is read in. `-180` is Brasília. |
| `--recurrence-horizon` | How many days ahead recurrences are written as pending transactions. The default is 31. |
| `--database` | The data folder. The default is `data`. |

To use the API, create a token on *API tokens*, then call it:

```bash
curl -X POST localhost:3000/api/admin/erp/dashboard -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

[Erp](docs/Erp/doc.md) explains the domain model. [ErpApi](docs/ErpApi/doc.md) explains the API, and [Routes](docs/Routes/doc.md) lists every call. `scripts/smoke_test.py` runs the whole system end to end.
