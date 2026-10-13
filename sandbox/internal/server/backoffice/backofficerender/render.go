package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Every backoffice page is rendered inside the layout every Convexa page
// shares — assets/backoffice/layout.html, with the navigation, the language
// switcher and the styles — by erppage. Templates run on html/template, so
// every value is escaped for the context it lands in.

// Notice is the outcome of the last add, edit or remove.
type Notice = erppage.Notice

// render answers, under status, the backoffice page at path for the request
// with props, titled title, the navigation item active marked, notice above
// it and body for its template.
func render(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, path string, title string, active string, notice Notice, body any) error {
	return erppage.Render(sandbox, response, status, props, path, title, active, notice, body)
}

// tr is text in the language the request with props is shown in.
func tr(sandbox *api.Sandbox, props *routeprops.RouteProps, text string) string {
	return erpi18n.Tr(sandbox, erppage.Language(sandbox, props), text)
}

// trf is format in the language the request with props is shown in, filled
// with args.
func trf(sandbox *api.Sandbox, props *routeprops.RouteProps, format string, args ...any) string {
	return erpi18n.Trf(sandbox, erppage.Language(sandbox, props), format, args...)
}

// initialOf is the first letter of name, "?" for an empty one.
func initialOf(sandbox *api.Sandbox, name string) string {
	for _, letter := range name {
		return string(letter)
	}
	return "?"
}
