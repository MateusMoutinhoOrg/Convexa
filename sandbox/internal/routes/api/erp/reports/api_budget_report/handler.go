package api_budget_report

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/budget-report: {"budgets": [...]}, every
// budget of the body's company and category whose period overlaps from to to,
// set against what its transactions hold.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	from, err := erpapi.Date(sandbox, props, "from", input.Body.From)
	if err != nil {
		return err
	}
	to, err := erpapi.Date(sandbox, props, "to", input.Body.To)
	if err != nil {
		return err
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lines := erpreports.ComputeBudgetLines(sandbox, book, erpstore.BudgetQuery{Company: int64(input.Body.Company), Category: int64(input.Body.Category), From: from, To: to})
	document := erpapi.Object(sandbox)
	document.AddItemToObject("budgets", erpapi.BudgetLinesJSON(sandbox, lines))
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
