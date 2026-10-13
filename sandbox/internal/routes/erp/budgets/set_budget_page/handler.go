package set_budget_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/manage/set-budget/{id}: the form that edits the
// budget, or the list when it no longer exists.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	budget, ok := erpstore.FindBudget(sandbox, int64(input.Id))
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-budgets", erpview.NoticeNotFound))
	}
	values := erpview.BudgetFormOf(sandbox, erppage.Language(sandbox, props), erpstore.BudgetFieldsOf(sandbox, budget))
	return erpview.RenderBudgetForm(sandbox, response, props, budget.Id, values, erpstore.Problem{})
}
