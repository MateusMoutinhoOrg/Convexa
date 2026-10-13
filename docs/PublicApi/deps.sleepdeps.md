# `deps.SleepDeps`

`sandbox/deps/sleepdeps`

## `Contract`

Contract is the waiting half of a clock, injected whole as the Deps.SleepDeps field. The loop that writes the occurrences of the recurrences is what reaches for it, to wait between two runs.

| Field | Type | Description |
| --- | --- | --- |
| `Sleep` | `func(milliseconds int64)` | Sleep pauses the calling goroutine for at least milliseconds. Zero or a negative value returns at once. |

[every contract](doc.md)
