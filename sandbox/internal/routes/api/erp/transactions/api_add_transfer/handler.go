package api_add_transfer

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-transfer: the transfer the body
// describes is added and answered under a 201 as {"out": {...}, "in": {...}},
// its two legs.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	date, err := erpapi.Date(sandbox, props, "date", input.Body.Date)
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
	fields := erpstore.TransferFields{
		FromAccount: int64(input.Body.FromAccount), ToAccount: int64(input.Body.ToAccount),
		FromCompany: int64(input.Body.FromCompany), ToCompany: int64(input.Body.ToCompany), Amount: int64(input.Body.Amount),
		Description: input.Body.Description, DocumentNumber: input.Body.DocumentNumber, Notes: input.Body.Notes,
		Date: date, DueDate: due, Status: input.Body.Status, SettledDate: settled, CardInvoice: invoice,
	}

	out, in, problem, err := erpstore.AddTransfer(sandbox, fields, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	today := erpdate.Today(sandbox)
	document := erpapi.Object(sandbox)
	document.AddItemToObject("out", erpapi.TransactionJSON(sandbox, out, today))
	document.AddItemToObject("in", erpapi.TransactionJSON(sandbox, in, today))
	return erpapi.Write(sandbox, response, api.StatusCreated, document)
}
