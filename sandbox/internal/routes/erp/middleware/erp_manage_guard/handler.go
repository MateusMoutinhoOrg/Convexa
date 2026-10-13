package erp_manage_guard

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficerender"
)

// Handle runs in front of every ANY /admin/manage/{*Rest}, after the
// backoffice-session-auth middleware put the signed-in user on props.User: it
// is a middleware. Every page and form that changes a financial record lives
// under /admin/manage, so a root or an admin declines, and the route after
// it runs; a basic user, who only reads, is answered the forbidden page under
// a 403.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	if !erppage.CanManage(sandbox, props) {
		return backofficerender.RenderForbiddenPageFor(sandbox, response, props, "Basic users can see every record and report, but only admin and root users can change them. Ask one of them if something needs to change.")
	}
	return nil
}
