# `GET /{*Rest}`

Serves any file of the embedded assets/front tree

## Try it

Only what is required:

```bash
curl localhost:3000/my-rest
```

With every value it reads:

```bash
curl 'localhost:3000/my-rest?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'accept-language: my-accept-language' \
  -b 'convexa_lang=my-convexa-lang'
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| `{*Rest}` | the rest of the address — one part or more, like `a/b.png` | `my-rest` | the file under assets/front; none is its index.html |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language to show the page or answer the call in: en or pt — read by [`erp-language`](erp_language.md), which runs first |
| `convexa_lang` | cookie | text | no | `my-convexa-lang` | the language a former ?lang= chose — read by [`erp-language`](erp_language.md), which runs first |
| `accept-language` | header | text | no | `my-accept-language` | read by [`erp-language`](erp_language.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html; charset=utf-8`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`api-get-current-backoffice-user`](api_get_current_backoffice_user.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-api-root-guard`](backoffice_api_root_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-api-token-auth`](backoffice_api_token_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-download-backup`](api_download_backup.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-download-backup-file`](api_download_backup_file.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-list-backup-files`](api_list_backup_files.md) | depends on the address — `explain-route` gives the exact answer |
| [`api-list-backups`](api_list_backups.md) | depends on the address — `explain-route` gives the exact answer |
| [`erp-api-manage-guard`](erp_api_manage_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-backoffice-api-token-page`](add_backoffice_api_token_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-home`](backoffice_home.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-index`](backoffice_index.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-login-page`](backoffice_login_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backoffice-api-tokens-page`](list_backoffice_api_tokens_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backoffice-users-page`](list_backoffice_users_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-root-guard`](backoffice_root_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-same-origin`](backoffice_same_origin.md) | depends on the address — `explain-route` gives the exact answer |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-backoffice-user-page`](add_backoffice_user_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`download-backup`](download_backup.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-backups-page`](list_backups_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-backoffice-user-page`](set_backoffice_user_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-account-page`](add_account_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-accounts-page`](list_accounts_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-account-page`](set_account_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-budget-page`](add_budget_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-budgets-page`](list_budgets_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-budget-page`](set_budget_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`card-invoice-page`](card_invoice_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-card-invoices-page`](list_card_invoices_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-credit-cards-page`](list_credit_cards_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-category-page`](add_category_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-categories-page`](list_categories_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-category-page`](set_category_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-company-page`](add_company_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-companies-page`](list_companies_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-company-page`](set_company_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`erp-language`](erp_language.md) | always |
| [`erp-manage-guard`](erp_manage_guard.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-recurrence-page`](add_recurrence_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-recurrences-page`](list_recurrences_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-recurrence-page`](set_recurrence_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`account-report-page`](account_report_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`budget-report-page`](budget_report_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`cash-flow-page`](cash_flow_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`company-report-page`](company_report_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`dashboard-page`](dashboard_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`payables-receivables-page`](payables_receivables_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`projection-page`](projection_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`settings-page`](settings_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-third-party-page`](add_third_party_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-third-parties-page`](list_third_parties_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-third-party-page`](set_third_party_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-transaction-page`](add_transaction_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`add-transfer-page`](add_transfer_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`list-transactions-page`](list_transactions_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-transaction-page`](set_transaction_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`set-transfer-page`](set_transfer_page.md) | depends on the address — `explain-route` gives the exact answer |
| [`health`](health.md) | depends on the address — `explain-route` gives the exact answer |
| [`openapi`](openapi.md) | depends on the address — `explain-route` gives the exact answer |

---

For developers: `sandbox/internal/routes/front/` · Assets · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
