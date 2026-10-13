package api_settle_transaction

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/settle-transaction: the transaction with the body's id — both legs of a transfer — is marked paid or received on settled-date, today when blank.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	date, err := erpapi.Date(sandbox, props, "settled-date", input.Body.SettledDate)
	if err != nil {
		return err
	}

	problem, err := erpstore.SettleTransaction(sandbox, int64(input.Body.Id), date)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.Ok(sandbox, response)
}
