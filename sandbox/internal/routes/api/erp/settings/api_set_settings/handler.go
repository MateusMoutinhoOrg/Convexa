package api_set_settings

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/root/set-settings: the settings are saved, for
// a root's token, and answered.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	problem, err := erpstore.WriteSettings(sandbox, erpstore.Settings{Currency: input.Body.Currency})
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("currency", erpstore.Currency(sandbox))
	document.AddItemToObject("utc-offset-minutes", sandbox.Config.UtcOffsetMinutes)
	document.AddItemToObject("recurrence-horizon-days", sandbox.Config.RecurrenceHorizonDays)
	document.AddItemToObject("today", erpdate.ISO(sandbox, erpdate.Today(sandbox)))
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
