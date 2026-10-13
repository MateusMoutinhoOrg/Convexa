package api_list_card_invoices

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-card-invoices: {"invoices": [...]},
// every invoice of the card with the body's id, oldest first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	card, ok := erpstore.FindAccount(sandbox, int64(input.Body.Id))
	if !ok || !erpledger.IsCard(sandbox, card) {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No credit card has that id.")
	}
	txs, err := erpstore.Transactions(sandbox)
	if err != nil {
		return err
	}
	invoices := erpapi.Array(sandbox)
	for _, invoice := range erpledger.Invoices(sandbox, card, txs, erpdate.Today(sandbox)) {
		invoices.AddItemToArray(erpapi.InvoiceJSON(sandbox, card.Id, invoice))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("invoices", invoices)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
