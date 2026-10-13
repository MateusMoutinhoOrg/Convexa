package api_get_budget

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/get-budget: {"budget": {...}}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	budget, ok := erpstore.FindBudget(sandbox, int64(input.Body.Id))
	if !ok {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No record has that id.")
	}
	return erpapi.One(sandbox, response, api.StatusOK, "budget", erpapi.BudgetJSON(sandbox, budget))
}
