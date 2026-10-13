package pay_card_invoice_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/pay-card-invoice/{id}: what is left of the
// invoice of the month posted is paid from the account posted, one transfer per
// company, and the browser sent back to the invoice.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	reader := erpview.Reader{}
	month := erpview.Month(sandbox, &reader, "month", input.Body.Month)
	date := erpview.Date(sandbox, &reader, "date", input.Body.Date)
	status := erpledger.StatusPending
	if input.Body.Settled != "" {
		status = erpledger.StatusSettled
	}
	back := erpview.Href(sandbox, "/admin/erp/card-invoice/"+sandbox.Deps.StringsDeps.FormatInt(int64(input.Id), 10), "month", input.Body.Month)
	if reader.Message != "" {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, back, erpview.NoticeRefused))
	}
	_, problem, err := erpstore.PayInvoice(sandbox, int64(input.Id), month, erpview.ParseId(sandbox, input.Body.FromAccount), date, status, props.User.Id)
	if err != nil {
		return err
	}
	notice := erpview.NoticePaid
	if problem.Status == api.StatusConflict {
		notice = erpview.NoticeNothingToPay
	} else if erpstore.Failed(sandbox, problem) {
		notice = erpview.NoticeOfProblem(sandbox, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, back, notice))
}
