package api_set_budget

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/set-budget: every field of the budget
// with the body's id is rewritten from the body.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	start, err := erpapi.Date(sandbox, props, "period-start", input.Body.PeriodStart)
	if err != nil {
		return err
	}
	end, err := erpapi.Date(sandbox, props, "period-end", input.Body.PeriodEnd)
	if err != nil {
		return err
	}
	fields := erpstore.BudgetFields{
		Company: int64(input.Body.Company), Category: int64(input.Body.Category), PeriodStart: start, PeriodEnd: end,
		Amount: int64(input.Body.Amount), Notes: input.Body.Notes, Months: int64(input.Body.Months),
	}

	budget, problem, err := erpstore.SetBudget(sandbox, int64(input.Body.Id), fields)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusOK, "budget", erpapi.BudgetJSON(sandbox, budget))
}
