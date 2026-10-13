# Routes

Every address this server answers. Open one to see what to send, a request you can run as it is,
and what comes back.

The requests call `localhost:3000`, where `convexa start-server` listens when that port is free —
it prints the address it took. Change it to wherever your server runs.

## How to read an address

| In the address | Means | For example |
| --- | --- | --- |
| `GET`, `POST`, … | the method to send it with; `ANY` takes every one | `curl -X POST …` |
| `/users` | exactly that text | `/users` |
| `{name}` | a value you choose | `/users/{tenant}` -> `/users/acme` |
| `{name:integer}` | a value of that type: `integer`, `number` or `uuid` | `/articles/{id:integer}` -> `/articles/42` |
| `{*name}` | the rest of the address, one part or more | `/files/{*file}` -> `/files/a/b.png` |
| `*` | anything else, or nothing | `/admin/*` -> `/admin`, `/admin/users` |
| `(a\|b)` | one of these words | `/(en\|pt)` -> `/en` |

## Backoffice Users API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/get-backoffice-user`](api_get_backoffice_user.md) | Answers one backoffice user by id |
| [`POST /api/admin/list-backoffice-users`](api_list_backoffice_users.md) | Lists backoffice users, filtered and paginated |
| [`POST /api/admin/root/add-backoffice-user`](api_add_backoffice_user.md) | Adds a backoffice user |
| [`POST /api/admin/root/remove-backoffice-user`](api_remove_backoffice_user.md) | Removes a backoffice user and every session of it |
| [`POST /api/admin/root/set-backoffice-user`](api_set_backoffice_user.md) | Edits a backoffice user; a missing or blank password keeps the current one |

## Backoffice API

| Route | What it does |
| --- | --- |
| [`GET /api/admin/me`](api_get_current_backoffice_user.md) | Answers the backoffice user of the Bearer token |

## Middleware

| Route | What it does |
| --- | --- |
| [`ANY /api/admin/root/*`](backoffice_api_root_guard.md) | Lets only root users reach /api/admin/root |
| [`ANY /api/admin/*`](backoffice_api_token_auth.md) | Requires a valid API token on every /api/admin route |
| [`ANY /api/admin/manage/*`](erp_api_manage_guard.md) | Lets only the tokens of root and admin users reach /api/admin/manage |
| [`ANY /*`](backoffice_client_ip.md) | Works out the client ip every route after it reads, from the connection or the reverse proxy in front |
| [`ANY /*`](backoffice_maintenance.md) | Answers 503 to every request while a restore writes the databases, and to every write while a snapshot reads them |
| [`ANY /admin/root/*`](backoffice_root_guard.md) | Lets only root users reach /admin/root |
| [`ANY /admin/*`](backoffice_same_origin.md) | Refuses a request to /admin another site's page sent |
| [`ANY /*`](backoffice_security_headers.md) | Sends the security headers on every /admin and /api/admin response, and the basic ones on every other |
| [`ANY /admin/* !(/admin/login)`](backoffice_session_auth.md) | Requires a valid backoffice session on /admin, except /admin/login |
| [`ANY /*`](erp_language.md) | Reads the language a request asks for: ?lang=, the convexa_lang cookie or Accept-Language |
| [`ANY /admin/manage/*`](erp_manage_guard.md) | Lets only root and admin users reach /admin/manage |

## Backoffice Backups API

| Route | What it does |
| --- | --- |
| [`POST /api/admin/root/add-backup-blob`](api_add_backup_blob.md) | Stores the body as one content of the backups, once whatever the number of snapshots that will name it, and answers its sha to reference it by |
| [`POST /api/admin/root/add-backup-file/{Id:integer}/{*Path}`](api_add_backup_file.md) | Stores the body as one file of an open snapshot, at the path below the --database folder the address ends with; a path it already holds is replaced |
| [`POST /api/admin/root/add-backup-reference`](api_add_backup_reference.md) | Adds to an open snapshot one file whose content add-backup-blob already stored, by its sha; a path it already holds is replaced |
| [`POST /api/admin/root/close-backup`](api_close_backup.md) | Closes an open snapshot once every content it names is stored: it turns ready, and can be downloaded and restored |
| [`POST /api/admin/root/create-backup`](api_create_backup.md) | Starts a snapshot of every database, under the name given or one after the current instant, and answers 202 at once: it is taken in the background |
| [`POST /api/admin/root/create-empty-backup`](api_create_empty_backup.md) | Records an empty open snapshot, under the name given or one after the current instant, that files are then added to one by one until it is closed |
| [`GET /api/admin/root/download-backup/{Id:integer}`](api_download_backup.md) | Downloads a snapshot as a zip archive |
| [`GET /api/admin/root/download-backup-file/{Id:integer}/{*Path}`](api_download_backup_file.md) | Downloads one file of a snapshot, at the path below the --database folder the address ends with |
| [`GET /api/admin/root/list-backup-files/{Id:integer}`](api_list_backup_files.md) | Lists every file of a snapshot, or the ones whose path starts with a prefix, as its path below the --database folder and the sha of its content |
| [`GET /api/admin/root/list-backups`](api_list_backups.md) | Lists every snapshot, or the ones whose name starts with a prefix, newest first, and whether a backup job is running |
| [`POST /api/admin/root/optimize-backup-storage`](api_optimize_backup_storage.md) | Starts removing every stored content no snapshot holds anymore, and answers 202 at once |
| [`POST /api/admin/root/remove-backup`](api_remove_backup.md) | Deletes a snapshot; the contents only it held stay stored until the backups size is optimized |
| [`POST /api/admin/root/restore-backup`](api_restore_backup.md) | Starts putting a snapshot back over every database, after a pre-restore snapshot of them, and answers 202 at once |
| [`POST /api/admin/root/upload-backup`](api_upload_backup.md) | Stores a zip archive a download built as a new snapshot |

## Accounts

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-account`](api_add_account.md) | Adds an account |
| [`POST /api/admin/erp/get-account`](api_get_account.md) | Answers one account |
| [`POST /api/admin/erp/list-accounts`](api_list_accounts.md) | Lists the accounts |
| [`POST /api/admin/manage/remove-account`](api_remove_account.md) | Removes an account |
| [`POST /api/admin/manage/set-account`](api_set_account.md) | Rewrites every field of an account |
| [`POST /admin/manage/add-account`](add_account_form.md) | Adds an account |
| [`GET /admin/manage/add-account`](add_account_page.md) | Shows the form that adds an account |
| [`GET /admin/erp/list-accounts`](list_accounts_page.md) | Lists the accounts |
| [`POST /admin/manage/remove-account/{Id:integer}`](remove_account_form.md) | Removes an account |
| [`POST /admin/manage/set-account/{Id:integer}`](set_account_form.md) | Edits an account |
| [`GET /admin/manage/set-account/{Id:integer}`](set_account_page.md) | Shows the form that edits an account |

## Budgets

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-budget`](api_add_budget.md) | Adds a budget, or one per month |
| [`POST /api/admin/erp/get-budget`](api_get_budget.md) | Answers one budget |
| [`POST /api/admin/erp/list-budgets`](api_list_budgets.md) | Lists the budgets |
| [`POST /api/admin/manage/remove-budget`](api_remove_budget.md) | Removes a budget |
| [`POST /api/admin/manage/set-budget`](api_set_budget.md) | Rewrites every field of a budget |
| [`POST /admin/manage/add-budget`](add_budget_form.md) | Adds a budget, or one per month |
| [`GET /admin/manage/add-budget`](add_budget_page.md) | Shows the form that adds budgets |
| [`GET /admin/erp/list-budgets`](list_budgets_page.md) | Lists the budgets |
| [`POST /admin/manage/remove-budget/{Id:integer}`](remove_budget_form.md) | Removes a budget |
| [`POST /admin/manage/set-budget/{Id:integer}`](set_budget_form.md) | Edits a budget |
| [`GET /admin/manage/set-budget/{Id:integer}`](set_budget_page.md) | Shows the form that edits a budget |

## Credit cards

| Route | What it does |
| --- | --- |
| [`POST /api/admin/erp/get-card-invoice`](api_get_card_invoice.md) | Answers one invoice of a credit card and its transactions |
| [`POST /api/admin/erp/list-card-invoices`](api_list_card_invoices.md) | Lists the invoices of a credit card |
| [`POST /api/admin/erp/list-credit-cards`](api_list_credit_cards.md) | Lists the credit cards and where they stand |
| [`POST /api/admin/manage/pay-card-invoice`](api_pay_card_invoice.md) | Pays what is left of a credit card invoice, one transfer per company |
| [`GET /admin/erp/card-invoice/{Id:integer}`](card_invoice_page.md) | Shows one invoice of a credit card |
| [`GET /admin/erp/list-card-invoices/{Id:integer}`](list_card_invoices_page.md) | Lists the invoices of a credit card |
| [`GET /admin/erp/list-credit-cards`](list_credit_cards_page.md) | Lists the credit cards and their current invoices |
| [`POST /admin/manage/pay-card-invoice/{Id:integer}`](pay_card_invoice_form.md) | Pays what is left of a credit card invoice |

## Categories

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-category`](api_add_category.md) | Adds a category |
| [`POST /api/admin/erp/get-category`](api_get_category.md) | Answers one category |
| [`POST /api/admin/erp/list-categories`](api_list_categories.md) | Lists the categories |
| [`POST /api/admin/manage/remove-category`](api_remove_category.md) | Removes a category |
| [`POST /api/admin/manage/set-category`](api_set_category.md) | Rewrites every field of a category |
| [`POST /admin/manage/add-category`](add_category_form.md) | Adds a category |
| [`GET /admin/manage/add-category`](add_category_page.md) | Shows the form that adds a category |
| [`GET /admin/erp/list-categories`](list_categories_page.md) | Lists the categories |
| [`POST /admin/manage/remove-category/{Id:integer}`](remove_category_form.md) | Removes a category |
| [`POST /admin/manage/set-category/{Id:integer}`](set_category_form.md) | Edits a category |
| [`GET /admin/manage/set-category/{Id:integer}`](set_category_page.md) | Shows the form that edits a category |

## Companies

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-company`](api_add_company.md) | Adds a company |
| [`POST /api/admin/erp/get-company`](api_get_company.md) | Answers one company |
| [`POST /api/admin/erp/list-companies`](api_list_companies.md) | Lists the companies |
| [`POST /api/admin/manage/remove-company`](api_remove_company.md) | Removes a company |
| [`POST /api/admin/manage/set-company`](api_set_company.md) | Rewrites every field of a company |
| [`POST /admin/manage/add-company`](add_company_form.md) | Adds a company |
| [`GET /admin/manage/add-company`](add_company_page.md) | Shows the form that adds a company |
| [`GET /admin/erp/list-companies`](list_companies_page.md) | Lists the companies |
| [`POST /admin/manage/remove-company/{Id:integer}`](remove_company_form.md) | Removes a company |
| [`POST /admin/manage/set-company/{Id:integer}`](set_company_form.md) | Edits a company |
| [`GET /admin/manage/set-company/{Id:integer}`](set_company_page.md) | Shows the form that edits a company |

## Recurrences

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-recurrence`](api_add_recurrence.md) | Adds a recurrence |
| [`POST /api/admin/manage/generate-recurrences`](api_generate_recurrences.md) | Writes the occurrences of every recurrence up to the horizon |
| [`POST /api/admin/erp/get-recurrence`](api_get_recurrence.md) | Answers one recurrence |
| [`POST /api/admin/erp/list-recurrences`](api_list_recurrences.md) | Lists the recurrences |
| [`POST /api/admin/manage/remove-recurrence`](api_remove_recurrence.md) | Removes a recurrence and its upcoming pending entries |
| [`POST /api/admin/manage/set-recurrence`](api_set_recurrence.md) | Rewrites every field of a recurrence |
| [`POST /admin/manage/add-recurrence`](add_recurrence_form.md) | Adds a recurrence |
| [`GET /admin/manage/add-recurrence`](add_recurrence_page.md) | Shows the form that adds a recurrence |
| [`POST /admin/manage/generate-recurrences`](generate_recurrences_form.md) | Writes the occurrences of every recurrence up to the horizon |
| [`GET /admin/erp/list-recurrences`](list_recurrences_page.md) | Lists the recurrences |
| [`POST /admin/manage/remove-recurrence/{Id:integer}`](remove_recurrence_form.md) | Removes a recurrence and its upcoming pending entries |
| [`POST /admin/manage/set-recurrence/{Id:integer}`](set_recurrence_form.md) | Edits a recurrence |
| [`GET /admin/manage/set-recurrence/{Id:integer}`](set_recurrence_page.md) | Shows the form that edits a recurrence |

## Reports

| Route | What it does |
| --- | --- |
| [`POST /api/admin/erp/account-balances`](api_account_balances.md) | Answers the balance of every account, whole and by company |
| [`POST /api/admin/erp/account-statement`](api_account_statement.md) | Answers the statement of an account |
| [`POST /api/admin/erp/budget-report`](api_budget_report.md) | Sets every budget against what happened |
| [`POST /api/admin/erp/cash-flow`](api_cash_flow.md) | Answers the cash flow of a period |
| [`POST /api/admin/erp/company-report`](api_company_report.md) | Answers what a company did over a period |
| [`POST /api/admin/erp/dashboard`](api_dashboard.md) | Answers the overview of every company's finances |
| [`POST /api/admin/erp/payables-receivables`](api_payables_receivables.md) | Lists what is left to pay and to receive |
| [`POST /api/admin/erp/projection`](api_projection.md) | Answers the projected balance, with and without the budgets |
| [`GET /admin/erp/account-report`](account_report_page.md) | Shows the balances of every account by company, and an account's statement |
| [`GET /admin/erp/budget-report`](budget_report_page.md) | Sets every budget against what happened |
| [`GET /admin/erp/cash-flow`](cash_flow_page.md) | Shows the cash flow of a period |
| [`GET /admin/erp/company-report`](company_report_page.md) | Shows what a company did over a period |
| [`GET /admin/erp/dashboard`](dashboard_page.md) | Shows the overview of every company's finances |
| [`GET /admin/erp/payables-receivables`](payables_receivables_page.md) | Lists what is left to pay and to receive |
| [`GET /admin/erp/projection`](projection_page.md) | Shows the projected balance, with and without the budgets |

## Settings

| Route | What it does |
| --- | --- |
| [`POST /api/admin/erp/get-settings`](api_get_settings.md) | Answers the settings |
| [`POST /api/admin/root/set-settings`](api_set_settings.md) | Saves the settings |
| [`POST /admin/root/settings`](settings_form.md) | Saves the settings |
| [`GET /admin/root/settings`](settings_page.md) | Shows the settings |

## Third parties

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-third-party`](api_add_third_party.md) | Adds a third party |
| [`POST /api/admin/erp/get-third-party`](api_get_third_party.md) | Answers one third party |
| [`POST /api/admin/erp/list-third-parties`](api_list_third_parties.md) | Lists the third parties |
| [`POST /api/admin/manage/remove-third-party`](api_remove_third_party.md) | Removes a third party |
| [`POST /api/admin/manage/set-third-party`](api_set_third_party.md) | Rewrites every field of a third party |
| [`POST /admin/manage/add-third-party`](add_third_party_form.md) | Adds a third party |
| [`GET /admin/manage/add-third-party`](add_third_party_page.md) | Shows the form that adds a third party |
| [`GET /admin/erp/list-third-parties`](list_third_parties_page.md) | Lists the third parties |
| [`POST /admin/manage/remove-third-party/{Id:integer}`](remove_third_party_form.md) | Removes a third party |
| [`POST /admin/manage/set-third-party/{Id:integer}`](set_third_party_form.md) | Edits a third party |
| [`GET /admin/manage/set-third-party/{Id:integer}`](set_third_party_page.md) | Shows the form that edits a third party |

## Transactions

| Route | What it does |
| --- | --- |
| [`POST /api/admin/manage/add-transaction`](api_add_transaction.md) | Adds a transaction, or a purchase in installments |
| [`POST /api/admin/manage/add-transfer`](api_add_transfer.md) | Adds a transfer between accounts or companies |
| [`POST /api/admin/erp/get-transaction`](api_get_transaction.md) | Answers one transaction |
| [`POST /api/admin/erp/list-transactions`](api_list_transactions.md) | Lists the transactions, newest first |
| [`POST /api/admin/manage/remove-transaction`](api_remove_transaction.md) | Removes a transaction, both legs of a transfer |
| [`POST /api/admin/manage/set-transaction`](api_set_transaction.md) | Rewrites every field of a transaction |
| [`POST /api/admin/manage/set-transfer`](api_set_transfer.md) | Rewrites both legs of a transfer |
| [`POST /api/admin/manage/settle-transaction`](api_settle_transaction.md) | Marks a transaction paid or received |
| [`POST /api/admin/manage/unsettle-transaction`](api_unsettle_transaction.md) | Marks a transaction pending again |
| [`POST /admin/manage/add-transaction`](add_transaction_form.md) | Adds a transaction, or a purchase in installments |
| [`GET /admin/manage/add-transaction`](add_transaction_page.md) | Shows the form that adds a transaction |
| [`POST /admin/manage/add-transfer`](add_transfer_form.md) | Adds a transfer between accounts or companies |
| [`GET /admin/manage/add-transfer`](add_transfer_page.md) | Shows the form that adds a transfer |
| [`GET /admin/erp/list-transactions`](list_transactions_page.md) | Lists the transactions |
| [`POST /admin/manage/remove-transaction/{Id:integer}`](remove_transaction_form.md) | Removes a transaction, both legs of a transfer |
| [`POST /admin/manage/set-transaction/{Id:integer}`](set_transaction_form.md) | Edits a transaction |
| [`GET /admin/manage/set-transaction/{Id:integer}`](set_transaction_page.md) | Shows the form that edits a transaction |
| [`POST /admin/manage/set-transfer/{Id:integer}`](set_transfer_form.md) | Edits a transfer |
| [`GET /admin/manage/set-transfer/{Id:integer}`](set_transfer_page.md) | Shows the form that edits a transfer |
| [`POST /admin/manage/settle-transaction/{Id:integer}`](settle_transaction_form.md) | Marks a transaction paid or received |
| [`POST /admin/manage/unsettle-transaction/{Id:integer}`](unsettle_transaction_form.md) | Marks a transaction pending again |

## Backoffice API Tokens

| Route | What it does |
| --- | --- |
| [`POST /admin/add-backoffice-api-token`](add_backoffice_api_token_form.md) | Creates an API token and shows it once |
| [`GET /admin/add-backoffice-api-token`](add_backoffice_api_token_page.md) | Shows the form that creates an API token |
| [`GET /admin/list-backoffice-api-tokens`](list_backoffice_api_tokens_page.md) | Lists your API tokens, or every user's for a root |
| [`POST /admin/revoke-backoffice-api-token/{Id:integer}`](revoke_backoffice_api_token_form.md) | Revokes an API token: your own, or anyone's for a root |

## Backoffice

| Route | What it does |
| --- | --- |
| [`GET /admin/home`](backoffice_home.md) | Shows the backoffice home page to the signed-in user |
| [`GET /admin`](backoffice_index.md) | Sends the signed-in user to the backoffice home page |
| [`POST /admin/login`](backoffice_login.md) | Signs a backoffice user in and sets the session cookie |
| [`GET /admin/login`](backoffice_login_page.md) | Shows the sign-in form |
| [`POST /admin/logout`](backoffice_logout.md) | Ends the current session |
| [`POST /admin/set-language`](set_language_form.md) | Saves the language the signed-in user sees the pages in |

## Backoffice Users

| Route | What it does |
| --- | --- |
| [`GET /admin/list-backoffice-users`](list_backoffice_users_page.md) | Lists backoffice users, filtered and paginated |
| [`POST /admin/root/add-backoffice-user`](add_backoffice_user_form.md) | Adds a backoffice user |
| [`GET /admin/root/add-backoffice-user`](add_backoffice_user_page.md) | Shows the form that adds a backoffice user |
| [`POST /admin/root/remove-backoffice-user/{Id:integer}`](remove_backoffice_user_form.md) | Removes a backoffice user and every session of it |
| [`POST /admin/root/set-backoffice-user/{Id:integer}`](set_backoffice_user_form.md) | Edits a backoffice user; a blank password keeps the current one |
| [`GET /admin/root/set-backoffice-user/{Id:integer}`](set_backoffice_user_page.md) | Shows the form that edits a backoffice user |

## Backoffice Backups

| Route | What it does |
| --- | --- |
| [`POST /admin/root/create-backup`](create_backup_form.md) | Starts a snapshot of every database, under the name the form gives or one after the current instant, and answers at once: it is taken in the background |
| [`GET /admin/root/download-backup/{Id:integer}`](download_backup.md) | Downloads a snapshot as a zip archive |
| [`GET /admin/root/list-backups`](list_backups_page.md) | Lists every snapshot of the databases, or the ones whose name starts with a prefix, with the controls that take, download, upload and restore one |
| [`POST /admin/root/optimize-backup-storage`](optimize_backup_storage_form.md) | Starts removing every stored content no snapshot holds anymore, and answers at once |
| [`POST /admin/root/remove-backup/{Id:integer}`](remove_backup_form.md) | Deletes a snapshot; the contents only it held stay stored until the backups size is optimized |
| [`POST /admin/root/restore-backup/{Id:integer}`](restore_backup_form.md) | Starts putting a snapshot back over every database, after a pre-restore snapshot of them, and answers at once |
| [`POST /admin/root/upload-backup`](upload_backup.md) | Stores a zip archive a download built as a new snapshot |

## Assets

| Route | What it does |
| --- | --- |
| [`GET /{*Rest}`](front.md) | Serves any file of the embedded assets/front tree |

## Server

| Route | What it does |
| --- | --- |
| [`GET /health`](health.md) | Reports that the server is up |
| [`GET /openapi.json`](openapi.md) | Answers the OpenAPI document of every route, to import into Postman or open in Swagger UI |

## Postman, Swagger and other tools

[openapi.json](openapi.json) holds every route below as an OpenAPI 3.0 document, and the running
server answers the same file at `/openapi.json`:

- **Postman**: *Import*, then pick the file or paste `http://localhost:3000/openapi.json`. Every
  route lands in a folder named after its section.
- **Swagger UI**, or any tool that reads OpenAPI: open `http://localhost:3000/openapi.json`.

A route that runs in front of others for every method — a check of a token, for example — is
not listed on its own there: what it reads is listed on each route it guards, and a token sent
as `Authorization` is the document's sign-in.

## When something goes wrong

| Status | Means |
| --- | --- |
| `400` | Something you sent is missing or has the wrong type or format |
| `401` | You have to identify yourself first — a token, for example |
| `403` | You are identified, but not allowed to do this |
| `404` | No route answers this address |
| `405` | The address exists, but not for this method — a `GET` where it takes a `POST`, for example |
| `413` | The body is too large |
| `415` | The body is not in the format the route reads — check `Content-Type` |
| `500` | The server failed while answering |

Unless the project changed it, the answer to an error is JSON naming what went wrong and, when
it is one value, which one:

```json
{"error": "required parameter 'authorization' is missing", "field": "authorization"}
```

For developers: each page, and `openapi.json`, is generated on every build from
`sandbox/internal/routes/<name>/route.yaml` ([RouteYaml](../RouteYaml/doc.md#openapi)); hidden
routes are left out. `agnos list-routes` prints the routes in the order they run, and
`agnos explain-route <METHOD> <path>` which ones a request reaches. The error answers are
the eight files of `sandbox/internal/server/errors/` ([RouteYaml](../RouteYaml/doc.md#failures)).
