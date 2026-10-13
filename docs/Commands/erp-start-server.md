# `erp-start-server`

Read the ERP settings in front of start-server and start writing the recurrences

A middleware: it runs on rung 51, in front of every command line matching
`start-server`, and hands the line on unless it answers.

Runs in front of start-server and answers nothing. It reads --utc-offset, the minutes from UTC the dates of Convexa are read in (-180 for Brasília), and --recurrence-horizon, how many days ahead the recurrences are written as pending transactions, then starts the job that writes them: now, and every 6 hours while the server runs.

| Flag | Type | Default | Description | From |
| --- | --- | --- | --- | --- |
| `--utc-offset` | integer | `0` | minutes from UTC the dates are read in: -180 for Brasília, 60 for Paris | — |
| `--recurrence-horizon` | integer | `31` | how many days ahead the occurrences of the recurrences are written as pending transactions | — |

| Runs in front of | When |
| --- | --- |
| [`start-server`](start-server.md) | always |

Middlewares · [every command](doc.md) · [CommandYaml](../CommandYaml/doc.md)
