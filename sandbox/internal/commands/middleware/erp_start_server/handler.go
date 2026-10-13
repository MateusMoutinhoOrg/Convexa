package erp_start_server

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpjobs"
)

// MaxUtcOffsetMinutes is the farthest from UTC a place's clock is: 14 hours.
const MaxUtcOffsetMinutes = 14 * 60

// MaxRecurrenceHorizonDays is the farthest ahead the recurrences are
// written: a year.
const MaxRecurrenceHorizonDays = 366

// Handle runs in front of `start-server` and answers nothing, so start-server
// runs after it. It puts the two flags Convexa adds to start-server on
// sandbox.Config — the minutes from UTC its dates are read in, and how many
// days ahead the recurrences are written — then starts the job that writes
// them, now and every few hours, for as long as the server runs.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	offset := int64(input.UtcOffset)
	if offset < -MaxUtcOffsetMinutes || offset > MaxUtcOffsetMinutes {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "utc-offset", sandbox.Deps.StdDeps.Sprintf("--utc-offset is in minutes, between -%d and %d: -180 for Brasília", MaxUtcOffsetMinutes, MaxUtcOffsetMinutes))
	}
	horizon := int64(input.RecurrenceHorizon)
	if horizon < 0 || horizon > MaxRecurrenceHorizonDays {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "recurrence-horizon", sandbox.Deps.StdDeps.Sprintf("--recurrence-horizon is in days, between 0 and %d", MaxRecurrenceHorizonDays))
	}
	sandbox.Config.UtcOffsetMinutes = offset
	sandbox.Config.RecurrenceHorizonDays = horizon
	erpjobs.StartRecurrences(sandbox)
	return nil
}
