package api_set_third_party

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/set-third-party: every field of the
// third party with the body's id is rewritten from the body — send the whole
// record, as get-third-party answers it — and the result answered.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	fields := erpstore.ThirdPartyFields{
		Name: input.Body.Name, Role: input.Body.Role, Document: input.Body.Document, Email: input.Body.Email,
		Phone: input.Body.Phone, Address: input.Body.Address, Notes: input.Body.Notes, Status: input.Body.Status,
	}
	record, problem, err := erpstore.SetThirdParty(sandbox, int64(input.Body.Id), fields)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusOK, "third-party", erpapi.ThirdPartyJSON(sandbox, record))
}
