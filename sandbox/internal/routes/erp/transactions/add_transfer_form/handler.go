package add_transfer_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/add-transfer: the transfer the form
// describes is added — its two legs — and the browser sent back to the list it
// came from; anything else answers the form again under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	values := erpview.TransferForm{
		FromAccount: input.Body.FromAccount, ToAccount: input.Body.ToAccount, FromCompany: input.Body.FromCompany,
		ToCompany: input.Body.ToCompany, Amount: input.Body.Amount, Description: input.Body.Description,
		DocumentNumber: input.Body.DocumentNumber, Notes: input.Body.Notes, Date: input.Body.Date, DueDate: input.Body.DueDate,
		Settled: input.Body.Settled, SettledDate: input.Body.SettledDate, CardInvoice: input.Body.CardInvoice, Next: input.Body.Next,
	}
	fields, problem := erpview.TransferFieldsOf(sandbox, values)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, _, problem, err = erpstore.AddTransfer(sandbox, fields, props.User.Id)
		if err != nil {
			return err
		}
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderTransferForm(sandbox, response, props, 0, values, problem)
	}
	next := erpview.SafeNext(sandbox, values.Next, "/admin/erp/list-transactions")
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, next, erpview.NoticeAdded))
}
