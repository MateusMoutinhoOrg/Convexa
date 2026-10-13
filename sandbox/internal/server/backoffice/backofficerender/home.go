package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// HomePage is what backoffice/home.html is rendered with: the signed-in
// user's own account.
type HomePage struct {
	Id             string
	SessionMinutes int
}

// RenderHomePage answers the account page of the signed-in user, whose
// session lasts sessionMinutes.
func RenderHomePage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, sessionMinutes int) error {
	return render(sandbox, response, api.StatusOK, props, "backoffice/home.html", "My account", "home", Notice{}, HomePage{
		Id:             sandbox.Deps.StringsDeps.FormatInt(props.User.Id, 10),
		SessionMinutes: sessionMinutes,
	})
}
