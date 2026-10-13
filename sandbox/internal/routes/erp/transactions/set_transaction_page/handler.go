package set_transaction_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/manage/set-transaction/{id}: the form that edits
// the transaction, filled with what it holds — the transfer form for a
// transfer leg — or the list, when it no longer exists.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	tx, ok := erpstore.FindTransaction(sandbox, int64(input.Id))
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-transactions", erpview.NoticeNotFound))
	}
	if erpledger.IsTransfer(sandbox, tx.Type) {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.Href(sandbox, "/admin/manage/set-transfer/"+sandbox.Deps.StringsDeps.FormatInt(tx.Id, 10), "next", input.Next))
	}
	values := erpview.FormOf(sandbox, erppage.Language(sandbox, props), erpstore.TransactionFieldsOf(sandbox, tx))
	values.Next = input.Next
	return erpview.RenderTransactionForm(sandbox, response, props, tx.Id, values, erpstore.Problem{})
}
