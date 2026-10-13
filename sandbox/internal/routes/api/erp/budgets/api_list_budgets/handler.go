package api_list_budgets

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-budgets: {"budgets": [...]}, every
// budget of the body's company and category whose period overlaps from to to.
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
	budgets, err := erpstore.Budgets(sandbox, erpstore.BudgetQuery{Company: int64(input.Body.Company), Category: int64(input.Body.Category), From: from, To: to})
	if err != nil {
		return err
	}
	list := erpapi.Array(sandbox)
	for _, budget := range budgets {
		list.AddItemToArray(erpapi.BudgetJSON(sandbox, budget))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("budgets", list)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
