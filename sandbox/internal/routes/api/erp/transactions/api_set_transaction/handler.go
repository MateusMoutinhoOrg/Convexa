package api_set_transaction

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/set-transaction: every field of the
// transaction with the body's id is rewritten from the body. A transfer leg is
// rewritten with set-transfer.
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

	tx, problem, err := erpstore.SetTransaction(sandbox, int64(input.Body.Id), fields, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusOK, "transaction", erpapi.TransactionJSON(sandbox, tx, erpdate.Today(sandbox)))
}
