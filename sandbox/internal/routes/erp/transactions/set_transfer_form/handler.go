package set_transfer_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/set-transfer/{id}: both legs of the
// transfer are rewritten from the form and the browser sent back to the list it
// came from; a form they cannot be written from is answered again under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	id := int64(input.Id)
	values := erpview.TransferForm{
		FromAccount: input.Body.FromAccount, ToAccount: input.Body.ToAccount, FromCompany: input.Body.FromCompany,
		ToCompany: input.Body.ToCompany, Amount: input.Body.Amount, Description: input.Body.Description,
		DocumentNumber: input.Body.DocumentNumber, Notes: input.Body.Notes, Date: input.Body.Date, DueDate: input.Body.DueDate,
		Settled: input.Body.Settled, SettledDate: input.Body.SettledDate, CardInvoice: input.Body.CardInvoice, Next: input.Body.Next,
	}
	fields, problem := erpview.TransferFieldsOf(sandbox, values)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, _, problem, err = erpstore.SetTransfer(sandbox, id, fields)
		if err != nil {
			return err
		}
	}
	next := erpview.SafeNext(sandbox, values.Next, "/admin/erp/list-transactions")
	if problem.Status == api.StatusNotFound {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, next, erpview.NoticeNotFound))
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderTransferForm(sandbox, response, props, id, values, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, next, erpview.NoticeUpdated))
}
