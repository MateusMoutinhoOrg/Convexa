package erp_api_manage_guard

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle runs in front of every ANY /api/admin/manage/{*Rest}, after the
// backoffice-api-token-auth middleware put the token's user on props.User: it
// is a middleware. Every call that changes a financial record lives under
// /api/admin/manage, so the token of a root or an admin declines, and the
// route after it runs; a basic user's is answered 403 in JSON.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	if !erppage.CanManage(sandbox, props) {
		return erpapi.Fail(sandbox, props, api.StatusForbidden, "", "only root and admin users may change the financial records")
	}
	return nil
}
