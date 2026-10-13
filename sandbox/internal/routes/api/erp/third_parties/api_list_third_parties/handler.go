package api_list_third_parties

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-third-parties: {"third-parties": [...]}, every
// third party the body's filters keep.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	records, err := erpstore.ThirdParties(sandbox, erpstore.ThirdPartyQuery{Search: input.Body.Search, Role: input.Body.Role, Status: input.Body.Status})
	if err != nil {
		return err
	}
	list := erpapi.Array(sandbox)
	for _, record := range records {
		list.AddItemToArray(erpapi.ThirdPartyJSON(sandbox, record))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("third-parties", list)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
