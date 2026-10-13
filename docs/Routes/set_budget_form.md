# `POST /admin/manage/set-budget/{Id:integer}`

Edits a budget

## Try it

Only what is required:

```bash
curl -X POST localhost:3000/admin/manage/set-budget/1 \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d ''
```

With every value it reads:

```bash
curl -X POST 'localhost:3000/admin/manage/set-budget/1?lang=my-lang' \
  -H 'x-client-ip: my-x-client-ip' \
  -H 'x-forwarded-for: my-x-forwarded-for' \
  -H 'x-request-method: my-x-request-method' \
  -H 'origin: my-origin' \
  -H 'host: my-host' \
  -H 'accept-language: my-accept-language' \
  -b 'convexa_lang=my-convexa-lang' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'amount=text&category=text&company=text&months=text&notes=text&period-end=text&period-start=text'
```

## In the address

| Part | What goes there | Example | Description |
| --- | --- | --- | --- |
| `{Id:integer}` | whole number | `1` |  |

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-client-ip` | header | text | no | `my-x-client-ip` | the ip of the connection, set by the server and never by the client — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-forwarded-for` | header | text | no | `my-x-forwarded-for` | the client chain a reverse proxy appended to, read only with --allow-x-forwarded-for — read by [`backoffice-client-ip`](backoffice_client_ip.md), which runs first |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client — read by [`backoffice-maintenance`](backoffice_maintenance.md), which runs first |
| `origin` | header | text | no | `my-origin` | the origin of the page that sent the request, sent by the browser — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `host` | header | text | no | `my-host` | the host the request was sent to — read by [`backoffice-same-origin`](backoffice_same_origin.md), which runs first |
| `lang` | query string | text | no | `my-lang` | the language to show the page or answer the call in: en or pt — read by [`erp-language`](erp_language.md), which runs first |
| `convexa_lang` | cookie | text | no | `my-convexa-lang` | the language a former ?lang= chose — read by [`erp-language`](erp_language.md), which runs first |
| `accept-language` | header | text | no | `my-accept-language` | read by [`erp-language`](erp_language.md), which runs first |

## Body

Send form fields, like `name=value&other=value` with the header `Content-Type: application/x-www-form-urlencoded`, up to 64 KB. The body is required.

| Field | What goes there | Required | Rules |
| --- | --- | --- | --- |
| `amount` | text | no |  |
| `category` | text | no |  |
| `company` | text | no |  |
| `months` | text | no |  |
| `notes` | text | no |  |
| `period-end` | text | no |  |
| `period-start` | text | no |  |

Example:

```text
amount=text&category=text&company=text&months=text&notes=text&period-end=text&period-start=text
```

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/html`. |
| `400` | Something you sent is missing or has the wrong type or format. The answer's `field` names it. |
| `404` | `{Id:integer}` is not a whole number, so this route does not answer the address. |
| `413` | The body is larger than 64 KB. |
| `415` | The body was not sent with `Content-Type: application/x-www-form-urlencoded`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`backoffice-client-ip`](backoffice_client_ip.md) | always |
| [`backoffice-maintenance`](backoffice_maintenance.md) | always |
| [`backoffice-same-origin`](backoffice_same_origin.md) | always |
| [`backoffice-security-headers`](backoffice_security_headers.md) | always |
| [`backoffice-session-auth`](backoffice_session_auth.md) | depends on the address — `explain-route` gives the exact answer |
| [`erp-language`](erp_language.md) | always |
| [`erp-manage-guard`](erp_manage_guard.md) | always |

---

For developers: `sandbox/internal/routes/set_budget_form/` · Budgets · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
