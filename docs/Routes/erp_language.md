# `ANY /*`

Reads the language a request asks for: ?lang=, the convexa_lang cookie or Accept-Language

## Try it

Only what is required:

```bash
curl localhost:3000/
```

With every value it reads:

```bash
curl 'localhost:3000/?lang=my-lang' \
  -H 'accept-language: my-accept-language' \
  -b 'convexa_lang=my-convexa-lang'
```

## Query string, headers and cookies

| Name | Sent in | What goes there | Required | Example | Description |
| --- | --- | --- | --- | --- | --- |
| `lang` | query string | text | no | `my-lang` | the language to show the page or answer the call in: en or pt |
| `convexa_lang` | cookie | text | no | `my-convexa-lang` | the language a former ?lang= chose |
| `accept-language` | header | text | no | `my-accept-language` |  |

## What comes back

| Status | Means |
| --- | --- |
| `200` | It worked. The answer comes as `text/plain`. |

Any route may also answer `404`, `405` or `500`: see [when something goes wrong](doc.md#when-something-goes-wrong).

---

For developers: `sandbox/internal/routes/erp_language/` · Middleware · [every route](doc.md) · [RouteYaml](../RouteYaml/doc.md)
