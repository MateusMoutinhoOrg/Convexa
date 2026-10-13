package api_get_category

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/get-category: {"category": {...}}, the category
// with the id the body names; 404 when none has it.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	record, ok := erpstore.FindCategory(sandbox, int64(input.Body.Id))
	if !ok {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No record has that id.")
	}
	all, err := erpstore.Categories(sandbox, erpstore.CategoryQuery{})
	if err != nil {
		return err
	}
	paths := erpstore.CategoryPaths(sandbox, all)
	return erpapi.One(sandbox, response, api.StatusOK, "category", erpapi.CategoryJSON(sandbox, record, paths[record.Id]))
}
