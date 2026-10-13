package backoffice_home

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/home with backoffice/home.html,
// rendered for the user the backoffice-session-auth middleware put on props.User.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	return backofficerender.RenderHomePage(sandbox, response, props, backofficeauth.SessionSeconds/60)
}
