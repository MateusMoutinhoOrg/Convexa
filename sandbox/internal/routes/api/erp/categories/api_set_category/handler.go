package api_set_category

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/set-category: every field of the
// category with the body's id is rewritten from the body — send the whole
// record, as get-category answers it — and the result answered.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	fields := erpstore.CategoryFields{Name: input.Body.Name, Type: input.Body.Type, Parent: int64(input.Body.Parent), Status: input.Body.Status}
	record, problem, err := erpstore.SetCategory(sandbox, int64(input.Body.Id), fields)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusOK, "category", erpapi.CategoryJSON(sandbox, record, record.Name))
}
