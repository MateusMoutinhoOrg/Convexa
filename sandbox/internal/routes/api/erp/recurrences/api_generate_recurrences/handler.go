package api_generate_recurrences

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/generate-recurrences: the occurrences
// of every active recurrence up to the horizon are written now, answered as
// {"written": n}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	written, err := erpstore.GenerateRecurrences(sandbox, props.User.Id)
	if err != nil {
		return err
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("written", written)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
