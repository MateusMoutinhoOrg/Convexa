# `ANY /*`

Answers 503 to every request while a restore writes the databases, and to every write while a snapshot reads them

## Try it

Only what is required:

```bash
curl localhost:3000/
```

With every value it reads:

```bash
curl 'localhost:3000/?lang=my-lang' \
  -H 'x-request-method: my-x-request-method' \
  -H 'accept-language: my-accept-language' \
  -b 'convexa_lang=my-convexa-lang'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `x-request-method` | header | text | no | `my-x-request-method` | the method of the request, set by the server and never by the client |
| `lang` | query string | text | no | `my-lang` | the language to show the page or answer the call in: en or pt — read by [`erp-language`](erp_language.md), which runs first |
| `convexa_lang` | cookie | text | no | `my-convexa-lang` | the language a former ?lang= chose — read by [`erp-language`](erp_language.md), which runs first |
| `accept-language` | header | text | no | `my-accept-language` | read by [`erp-language`](erp_language.md), which runs first |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `application/json`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

## Runs first

These routes run before this one, on the same request. Any of them may refuse it — with `401`
or `403`, for example — or let it through to this route.

| Route | When |
| --- | --- |
| [`erp-language`](erp_language.md) | always |

---

For developers: `sandbox/internal/routes/backoffice_maintenance/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
