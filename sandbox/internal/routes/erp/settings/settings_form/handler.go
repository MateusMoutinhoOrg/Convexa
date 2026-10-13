package settings_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/root/settings: the settings are saved and the page
// answered again; a value that cannot be saved is answered under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	settings := erpstore.Settings{Currency: input.Body.Currency}
	problem, err := erpstore.WriteSettings(sandbox, settings)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderSettingsPage(sandbox, response, props, settings, problem, "")
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/root/settings", erpview.NoticeUpdated))
}
