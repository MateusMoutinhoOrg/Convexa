package erp_language

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// CookieName is the cookie a ?lang= leaves behind, so the next page — the
// sign-in page above all, which has no user to read a choice from — keeps the
// language without asking again.
const CookieName = "convexa_lang"

// CookieSeconds is how long the language cookie lasts: a year.
const CookieSeconds = 365 * 24 * 60 * 60

// Handle runs in front of every request, before anything that may answer a
// page: it is a middleware. It puts on props the language the request asked
// for — ?lang= first, else the cookie a former ?lang= left — and the one its
// Accept-Language header prefers, and the path the request named; a ?lang=
// Convexa speaks is remembered in the cookie. It always declines.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	props.Path = input.FullRoute
	props.AcceptLanguage = erpi18n.FromAcceptLanguage(sandbox, input.AcceptLanguage)
	if lang := erpi18n.Normalize(sandbox, input.Lang); lang != "" {
		props.Language = lang
		response.AddHeader("Set-Cookie", sandbox.Deps.StdDeps.Sprintf("%s=%s; Path=/; Max-Age=%d; SameSite=Lax", CookieName, lang, CookieSeconds))
		return nil
	}
	props.Language = erpi18n.Normalize(sandbox, input.ConvexaLang)
	return nil
}
