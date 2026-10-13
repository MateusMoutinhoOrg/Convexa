# Convexa

[![Go Reference](https://pkg.go.dev/badge/github.com/MateusMoutinhoOrg/Convexa.svg)](https://pkg.go.dev/github.com/MateusMoutinhoOrg/Convexa)
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


## Documentation

### CliUsage

Driving the CLI from a terminal - install, commands, flags, exit codes

| Doc | Description |
| --- | --- |
| [CliInstall](docs/CliInstall/doc.md) | Install convexa: a released binary per platform, or a build from source |
| [Commands](docs/Commands/doc.md) | Every command of convexa, generated from the command declarations |
| [CliExamples](docs/CliExamples/doc.md) | Index of every runnable example of the convexa cli |

### ServerUsage

Serving http - bring the server up, declare routes, read a body

| Doc | Description |
| --- | --- |
| [ServerUsage](docs/ServerUsage/doc.md) | Serve http from convexa: bring the layer up, declare routes, read a body |
| [FrontUsage](docs/FrontUsage/doc.md) | Serve a website from convexa: bring the front layer up, drop files in assets/front, add pages |
| [Routes](docs/Routes/doc.md) | Every address convexa answers: what to send, a request to run, what comes back, and the OpenAPI document for Postman or Swagger |
| [Backoffice](docs/Backoffice/doc.md) | The admin backoffice of convexa: secret, users, roles, sessions, API tokens and the /api/admin JSON api |
| [Backups](docs/Backups/doc.md) | Backups of every database of convexa from the backoffice: take a snapshot in the background, download, upload and restore one, or build one by hand file by file |

### LibUsage

Using the project as a Go module - wiring the deps, calling the sandbox

| Doc | Description |
| --- | --- |
| [LibUsage](docs/LibUsage/doc.md) | Use convexa as a Go module: wire the deps, build the sandbox, call its API |
| [PublicApi](docs/PublicApi/doc.md) | Every exported symbol of convexa, generated from the contract sources and their doc comments |
| [LibExamples](docs/LibExamples/doc.md) | Index of every runnable example of convexa as a Go module |

### ApiUsage

The exported surface - contracts, deps and what each one promises

| Doc | Description |
| --- | --- |
| [Erp Api](docs/ErpApi/doc.md) | The JSON api of the ERP: tokens, roles, conventions, every call and examples |

### Architecture

How the project is put together - layers, boundaries, data flow

| Doc | Description |
| --- | --- |
| [Databases](docs/Databases/doc.md) | Every database of convexa, generated from the table declarations |
| [Adapters](docs/Adapters/doc.md) | Contract, adapter and binding: three units, one field of Deps, and who fills it |
| [Erp](docs/Erp/doc.md) | The financial model of Convexa: companies, accounts, transactions, cards, recurrences, budgets, reports and roles |

### Development

Changing this repository - schema, build mechanics, recipes

| Doc | Description |
| --- | --- |
| [Requirements](docs/Requirements/doc.md) | The two tools this project needs — Go and agnos — installed per platform |
| [Workflow](docs/Workflow/doc.md) | Every change this project takes and the agnos command that makes it |
| [Rules](docs/Rules/doc.md) | Every rule the generators, `verify` and the hand-written files must hold to |
| [Structure](docs/Structure/doc.md) | The project schema: what lives where, what is generated, what verify enforces |

### Reference

Lookup tables - schemas, file formats, generated file listings

| Doc | Description |
| --- | --- |
| [CommandYaml](docs/CommandYaml/doc.md) | Every key of a command's command.yaml and what the generated code does with it |
| [Extensions](docs/Extensions/doc.md) | The generation mechanics this project turns on, and what each one writes |
| [DepList](docs/DepList/doc.md) | Every dep `agnos add-dep` can add, the adapters that fill it, and what backs each one |
| [Databases](docs/Databases/doc.md) | Every database of convexa, generated from the table declarations |
| [GeneratedFiles](docs/GeneratedFiles/doc.md) | Every file agnos writes into this project and whether build overwrites it |
| [RouteYaml](docs/RouteYaml/doc.md) | Every key of a route's route.yaml and what the generated code does with it |
| [Languages](docs/Languages/doc.md) | English first, Portuguese on request: how pages and the api pick a language, and how a text is translated |

## License

MIT License

Copyright (c) 2026 MateusMoutinho

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

