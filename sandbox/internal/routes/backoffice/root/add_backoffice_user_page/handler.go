package add_backoffice_user_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeusers"
)

// Handle answers GET /admin/root/add-backoffice-user with the
// empty form that POST /admin/root/add-backoffice-user reads, the viewer role
// selected.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	fields := backofficeusers.Fields{Role: int64(backofficeauth.RoleBasic)}
	return backofficerender.RenderAddUserPage(sandbox, response, api.StatusOK, props, fields, "")
}
