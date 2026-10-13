package api_get_card_invoice

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/get-card-invoice: {"invoice": {...},
// "shares": [...], "transactions": [...]}, the invoice of the body's month —
// the open one when blank — of the card with the body's id.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	card, ok := erpstore.FindAccount(sandbox, int64(input.Body.Id))
	if !ok || !erpledger.IsCard(sandbox, card) {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No credit card has that id.")
	}
	month, err := erpapi.Month(sandbox, props, "month", input.Body.Month)
	if err != nil {
		return err
	}
	today := erpdate.Today(sandbox)
	if month == 0 {
		month = erpledger.InvoiceFor(sandbox, card, today)
	}
	txs, err := erpstore.Transactions(sandbox)
	if err != nil {
		return err
	}
	shares := erpapi.Array(sandbox)
	for _, share := range erpledger.InvoiceShares(sandbox, card, txs, month) {
		object := erpapi.Object(sandbox)
		object.AddItemToObject("company", share.Company)
		object.AddItemToObject("charges", share.Charges)
		object.AddItemToObject("paid", share.Paid)
		object.AddItemToObject("remaining", share.Remaining)
		shares.AddItemToArray(object)
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("invoice", erpapi.InvoiceJSON(sandbox, card.Id, erpledger.FindInvoice(sandbox, card, txs, month, today)))
	document.AddItemToObject("shares", shares)
	document.AddItemToObject("transactions", erpapi.TransactionsJSON(sandbox, erpstore.FilterTransactions(sandbox, txs, nil, erpstore.TransactionQuery{Account: card.Id, CardInvoice: month}), today))
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
