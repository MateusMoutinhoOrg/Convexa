package api_get_transaction

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/get-transaction: {"transaction": {...}},
// the one with the body's id — and, for a transfer leg, "counterpart-transaction",
// its other leg.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	tx, ok := erpstore.FindTransaction(sandbox, int64(input.Body.Id))
	if !ok {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No record has that id.")
	}
	today := erpdate.Today(sandbox)
	document := erpapi.Object(sandbox)
	document.AddItemToObject("transaction", erpapi.TransactionJSON(sandbox, tx, today))
	if other, ok := erpstore.FindTransaction(sandbox, tx.Counterpart); ok && tx.Counterpart != 0 {
		document.AddItemToObject("counterpart-transaction", erpapi.TransactionJSON(sandbox, other, today))
	}
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
