package api_add_transaction

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-transaction: the transaction the
// body describes — or, with installments above 1, that many a month apart — is
// added and answered under a 201 as {"transactions": [...]}; status is pending
// or settled, amount is in cents, positive unless the type is adjustment.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	from, err := erpapi.Date(sandbox, props, "date", input.Body.Date)
	if err != nil {
		return err
	}
	due, err := erpapi.Date(sandbox, props, "due-date", input.Body.DueDate)
	if err != nil {
		return err
	}
	settled, err := erpapi.Date(sandbox, props, "settled-date", input.Body.SettledDate)
	if err != nil {
		return err
	}
	invoice, err := erpapi.Month(sandbox, props, "card-invoice", input.Body.CardInvoice)
	if err != nil {
		return err
	}
	fields := erpstore.TransactionFields{
		Company: int64(input.Body.Company), Account: int64(input.Body.Account), Category: int64(input.Body.Category),
		ThirdParty: int64(input.Body.ThirdParty), Type: input.Body.Type, Amount: int64(input.Body.Amount),
		Description: input.Body.Description, DocumentNumber: input.Body.DocumentNumber, Notes: input.Body.Notes,
		Date: from, DueDate: due, Status: input.Body.Status, SettledDate: settled, CardInvoice: invoice,
		Installments: int64(input.Body.Installments),
	}

	created, problem, err := erpstore.AddTransaction(sandbox, fields, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("transactions", erpapi.TransactionsJSON(sandbox, created, erpdate.Today(sandbox)))
	return erpapi.Write(sandbox, response, api.StatusCreated, document)
}
