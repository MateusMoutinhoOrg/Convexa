package settle_transaction_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/settle-transaction/{id}: the transaction — both legs of a transfer — is marked paid or received on the date posted, today when blank, and the browser sent back to the
// page it came from, told how it went.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	reader := erpview.Reader{}
	date := erpview.Date(sandbox, &reader, "settled-date", input.Body.SettledDate)

	problem, err := erpstore.SettleTransaction(sandbox, int64(input.Id), date)
	if err != nil {
		return err
	}
	notice := erpview.NoticeSettled
	if erpstore.Failed(sandbox, problem) {
		notice = erpview.NoticeOfProblem(sandbox, problem)
	}
	next := erpview.SafeNext(sandbox, input.Body.Next, "/admin/erp/list-transactions")
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, next, notice))
}
