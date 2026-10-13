package set_language_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routes/erp/middleware/erp_language"
)

// DefaultNext is where the browser goes back to when the form names no page
// of the backoffice to return to.
const DefaultNext = "/admin/erp/dashboard"

// Handle answers POST /admin/set-language: the signed-in user's language is
// saved on their record — so it follows them to any browser — and in the
// language cookie, and the browser is sent back to the page it came from. Only
// a path of the backoffice is returned to, never another site.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	lang := erpi18n.Normalize(sandbox, input.Body.Language)
	if lang == "" {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "language", "language must be en or pt")
	}
	err := backoffice_db.New(sandbox).SetBackofficeUserLanguage(props.User.Id, lang)
	if err != nil {
		return err
	}
	response.AddHeader("Set-Cookie", sandbox.Deps.StdDeps.Sprintf("%s=%s; Path=/; Max-Age=%d; SameSite=Lax", erp_language.CookieName, lang, erp_language.CookieSeconds))

	next := input.Body.Next
	if !sandbox.Deps.StringsDeps.HasPrefix(next, "/admin") || sandbox.Deps.StringsDeps.HasPrefix(next, "//") || sandbox.Deps.StringsDeps.ContainsAny(next, "\\\r\n") || next == "/admin/set-language" {
		next = DefaultNext
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, next)
}
