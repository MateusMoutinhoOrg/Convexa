# `deps.PageDeps`

`sandbox/deps/pagedeps`

## `Contract`

Contract is the html page library injected whole as the Deps.PageDeps field.

| Field | Type | Description |
| --- | --- | --- |
| `RenderHTMLPage` | `func(paths []string, entry string, vars interface{}) ([]byte, error)` | RenderHTMLPage parses every template of paths, in order, into one set, executes the template named entry with vars and returns the result. Every value is escaped for the context it lands in — text, an attribute, a url — so a value nobody escaped by hand can never become markup. The error reports a template that is missing, does not parse, or fails while executing. |

[every contract](doc.md)
