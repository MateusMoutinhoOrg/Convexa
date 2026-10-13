package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// ForbiddenPage is what backoffice/forbidden.html is rendered with.
type ForbiddenPage struct {
	// Reason says which role the page needs.
	Reason string
}

// RenderForbiddenPage answers, under a 403, the page telling the signed-in
// user that what they asked for needs the root role.
func RenderForbiddenPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps) error {
	return RenderForbiddenPageFor(sandbox, response, props, "Only root users can manage users, backups and settings. Ask a root user if something needs to change.")
}

// RenderForbiddenPageFor answers, under a 403, the page telling the signed-in
// user that what they asked for is not for their role, reason saying why.
func RenderForbiddenPageFor(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, reason string) error {
	return render(sandbox, response, api.StatusForbidden, props, "backoffice/forbidden.html", "Forbidden", "", Notice{}, ForbiddenPage{Reason: tr(sandbox, props, reason)})
}
