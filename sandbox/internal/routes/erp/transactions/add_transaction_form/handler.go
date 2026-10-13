package add_transaction_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/add-transaction: the transaction — or the
// installments — the form describes is added and the browser sent back to the
// list it came from; anything else answers the form again under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	values := erpview.TransactionForm{
		Company: input.Body.Company, Account: input.Body.Account, Category: input.Body.Category, ThirdParty: input.Body.ThirdParty,
		Type: input.Body.Type, Amount: input.Body.Amount, Description: input.Body.Description, DocumentNumber: input.Body.DocumentNumber,
		Notes: input.Body.Notes, Date: input.Body.Date, DueDate: input.Body.DueDate, Settled: input.Body.Settled,
		SettledDate: input.Body.SettledDate, CardInvoice: input.Body.CardInvoice, Installments: input.Body.Installments, Next: input.Body.Next,
	}
	fields, problem := erpview.FieldsOf(sandbox, values)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, problem, err = erpstore.AddTransaction(sandbox, fields, props.User.Id)
		if err != nil {
			return err
		}
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderTransactionForm(sandbox, response, props, 0, values, problem)
	}
	next := erpview.SafeNext(sandbox, values.Next, "/admin/erp/list-transactions")
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, next, erpview.NoticeAdded))
}
