package api_add_recurrence

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-recurrence: the recurrence is
// written from the body — its occurrences up to the horizon as pending
// transactions — and answered as {"recurrence": {...}}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	start, err := erpapi.Date(sandbox, props, "start-date", input.Body.StartDate)
	if err != nil {
		return err
	}
	end, err := erpapi.Date(sandbox, props, "end-date", input.Body.EndDate)
	if err != nil {
		return err
	}
	fields := erpstore.RecurrenceFields{
		Company: int64(input.Body.Company), Account: int64(input.Body.Account), Category: int64(input.Body.Category),
		ThirdParty: int64(input.Body.ThirdParty), Type: input.Body.Type, Amount: int64(input.Body.Amount),
		Description: input.Body.Description, Notes: input.Body.Notes, Frequency: input.Body.Frequency,
		Interval: int64(input.Body.Interval), StartDate: start, EndDate: end, Status: input.Body.Status,
	}

	recurrence, problem, err := erpstore.AddRecurrence(sandbox, fields, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusCreated, "recurrence", erpapi.RecurrenceJSON(sandbox, recurrence, erpstore.NextOccurrence(sandbox, recurrence, erpdate.AddDays(sandbox, erpdate.Today(sandbox), -1))))
}
