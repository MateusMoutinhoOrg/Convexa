# Languages

Convexa is written in English: the code, the API keys and enum values, and every text. Portuguese is offered on top of that.

**Pages.** Each page is shown in one language, picked in this order:
1. the language the signed-in user chose with the **EN / PT** switch, saved on their user record;
2. the `convexa_lang` cookie, which a `?lang=pt` leaves behind (this is what the sign-in page uses);
3. the browser's `Accept-Language`;
4. English.

**API.** Messages follow `?lang=` or `Accept-Language` first, then the token's user's choice. JSON keys and enum values never change with the language.

**The catalog.** `sandbox/internal/erp/erpi18n/pt.go` maps each English text to its Portuguese one.
- A template asks for a translation with `{{call .T "Save"}}`, or `{{call .Tf "%d of %d" .From .To}}` when it fills in values.
- Go code asks with `erpi18n.Tr` and `erpi18n.Trf`.
- A text with no entry is shown in English.
- A message the backoffice formatted already, such as "The password needs at least 12 characters.", is matched against its format, so it is translated too.

**Formats.** Amounts and dates are written the way each language writes them:

| | English | Portuguese |
|---|---|---|
| Amount | `1,234.56` | `1.234,56` |
| Date | `Oct 12, 2026` | `12/10/2026` |

Amount inputs accept either spelling.

**Adding a text.** Write it in English. Then add its entry to `pt.go`.
