package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// LoginPage is what backoffice/login.html is rendered with.
type LoginPage struct {
	// Error is shown above the form, "" for none.
	Error string
	// Username refills the login field after a failed attempt.
	Username string
}

// RenderLoginPage answers the login page under status, with message above the form.
func RenderLoginPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, status int, message string, username string) error {
	return render(sandbox, response, status, props, "backoffice/login.html", "Sign in", "", Notice{}, LoginPage{Error: tr(sandbox, props, message), Username: username})
}
