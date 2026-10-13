package erpjobs

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// The jobs of Convexa run on goroutines of their own while the server runs:
// for now one, which keeps the occurrences of the recurrences written as
// pending transactions up to the horizon, so the payables and receivables
// list them before they are due without anyone pressing a button.

// RecurrenceIntervalMilliseconds is how long the recurrence job waits between
// two runs: six hours, so a day never goes by without one.
const RecurrenceIntervalMilliseconds = 6 * 60 * 60 * 1000

// started holds a token once the recurrence job runs, so a second start is
// a no-op.
var started = make(chan struct{}, 1)

// StartRecurrences writes the occurrences of every active recurrence up to
// the horizon now, then every RecurrenceIntervalMilliseconds, on a goroutine
// of its own, for as long as the process runs. A failed run is reported on
// stderr and tried again at the next one. A second call does nothing.
func StartRecurrences(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	go func() {
		for {
			written, err := erpstore.GenerateRecurrences(sandbox, 0)
			if err != nil {
				sandbox.Deps.StdDeps.Eprintf("warning: writing the occurrences of the recurrences failed, it is tried again in 6 hours: %s\n", err.Error())
			} else if written > 0 {
				sandbox.Deps.StdDeps.Logf("recurrences: %d pending transactions written up to %d days ahead\n", written, sandbox.Config.RecurrenceHorizonDays)
			}
			sandbox.Deps.SleepDeps.Sleep(RecurrenceIntervalMilliseconds)
		}
	}()
}
