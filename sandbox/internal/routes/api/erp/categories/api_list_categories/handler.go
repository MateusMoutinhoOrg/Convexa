package api_list_categories

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-categories: {"categories": [...]}, every
// category the body's filters keep.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	records, err := erpstore.Categories(sandbox, erpstore.CategoryQuery{Search: input.Body.Search, Type: input.Body.Type, Status: input.Body.Status})
	if err != nil {
		return err
	}
	all, err := erpstore.Categories(sandbox, erpstore.CategoryQuery{})
	if err != nil {
		return err
	}
	paths := erpstore.CategoryPaths(sandbox, all)
	list := erpapi.Array(sandbox)
	for _, record := range records {
		list.AddItemToArray(erpapi.CategoryJSON(sandbox, record, paths[record.Id]))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("categories", list)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
