package api_add_budget

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-budget: the budget the body
// describes — or, with months above 0, one per month from the month of
// period-start — is added and answered under a 201 as {"budgets": [...]}.
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

	budgets, problem, err := erpstore.AddBudgets(sandbox, fields)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	list := erpapi.Array(sandbox)
	for _, budget := range budgets {
		list.AddItemToArray(erpapi.BudgetJSON(sandbox, budget))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("budgets", list)
	return erpapi.Write(sandbox, response, api.StatusCreated, document)
}
