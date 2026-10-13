package backoffice_root_guard

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficerender"
)

// Handle runs in front of every ANY /admin/root/{*Rest}, after the
// backoffice-session-auth middleware put the signed-in user on props.User: it is a
// middleware. A root declines, so the route after it runs; anyone else is
// answered the forbidden page under a 403.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	if backofficeauth.Role(props.User.Role) != backofficeauth.RoleRoot {
		return backofficerender.RenderForbiddenPage(sandbox, response, props)
	}
	return nil
}
