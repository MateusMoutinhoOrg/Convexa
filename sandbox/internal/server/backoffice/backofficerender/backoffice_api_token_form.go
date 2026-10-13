package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeapitokens"
)

// BackofficeApiTokenFormPage is what backoffice/backoffice_api_token_form.html
// is rendered with: the form that creates an API token.
type BackofficeApiTokenFormPage struct {
	// Error is shown above the form, "" for none.
	Error       string
	Name        string
	Expirations []ExpirationOption
	// IsCustom shows the date field, for the custom expiration.
	IsCustom bool
	Date     string
	// MinDate is the first day the date field may name: today, in UTC.
	MinDate string
	Ips     string
	// ClientIp is the ip the browser's request came from, offered as one
	// click to fill the ips field.
	ClientIp string
}

// ExpirationOption is one expiration the form offers.
type ExpirationOption struct {
	Value    string
	Label    string
	Selected bool
}

// expirationOptions are every expiration, the one of fields selected.
func expirationOptions(sandbox *api.Sandbox, props *routeprops.RouteProps, fields backofficeapitokens.Fields) []ExpirationOption {
	options := []ExpirationOption{}
	for _, expiration := range backofficeapitokens.Expirations(sandbox) {
		options = append(options, ExpirationOption{
			Value:    expiration.Value,
			Label:    tr(sandbox, props, expiration.Label),
			Selected: expiration.Value == fields.Expiration,
		})
	}
	return options
}

// RenderAddApiTokenPage answers, under status, the form that creates
// an API token for user, filled with fields and with message above it, the
// client ip clientIp offered for the ips field.
func RenderAddApiTokenPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, fields backofficeapitokens.Fields, clientIp string, message string) error {
	now := sandbox.Deps.StdDeps.Now() / 1_000_000_000
	return render(sandbox, response, status, props, "backoffice/backoffice_api_token_form.html", "New API token", "tokens", Notice{}, BackofficeApiTokenFormPage{
		Error:       tr(sandbox, props, message),
		Name:        fields.Name,
		Expirations: expirationOptions(sandbox, props, fields),
		IsCustom:    fields.Expiration == backofficeapitokens.ExpirationCustom,
		Date:        fields.Date,
		MinDate:     sandbox.Deps.TimeDeps.FormatUnix(now, backofficeapitokens.DateLayout),
		Ips:         fields.Ips,
		ClientIp:    clientIp,
	})
}
