package set_account_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/manage/set-account/{id}: the form that edits the
// account, filled with what it holds; the list, when it no longer exists.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	record, ok := erpstore.FindAccount(sandbox, int64(input.Id))
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-accounts", erpview.NoticeNotFound))
	}
	return erpview.RenderAccountForm(sandbox, response, props, record.Id, erpstore.AccountFieldsOf(sandbox, record), erpstore.Problem{})
}
