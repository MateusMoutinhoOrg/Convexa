package api_pay_card_invoice

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/pay-card-invoice: what is left of the
// invoice of the body's month of the card with the body's id is paid from
// from-account — the card's payment account when 0 — on date, settled or
// pending as status says: one transfer per company, answered as
// {"paid", "transfers": [...]}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	month, err := erpapi.Month(sandbox, props, "month", input.Body.Month)
	if err != nil {
		return err
	}
	date, err := erpapi.Date(sandbox, props, "date", input.Body.Date)
	if err != nil {
		return err
	}
	payment, problem, err := erpstore.PayInvoice(sandbox, int64(input.Body.Id), month, int64(input.Body.FromAccount), date, input.Body.Status, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("paid", payment.Paid)
	document.AddItemToObject("transfers", erpapi.TransactionsJSON(sandbox, payment.Transfers, erpdate.Today(sandbox)))
	return erpapi.Write(sandbox, response, api.StatusCreated, document)
}
