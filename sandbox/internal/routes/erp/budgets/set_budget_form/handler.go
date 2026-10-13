package set_budget_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/set-budget/{id}: the budget is rewritten
// from the form and the browser sent to the list.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	id := int64(input.Id)
	values := erpview.BudgetForm{
		Company: input.Body.Company, Category: input.Body.Category, PeriodStart: input.Body.PeriodStart,
		PeriodEnd: input.Body.PeriodEnd, Amount: input.Body.Amount, Notes: input.Body.Notes, Months: input.Body.Months,
	}
	fields, problem := erpview.BudgetFieldsOf(sandbox, values)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, problem, err = erpstore.SetBudget(sandbox, id, fields)
		if err != nil {
			return err
		}
	}
	if problem.Status == api.StatusNotFound {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-budgets", erpview.NoticeNotFound))
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderBudgetForm(sandbox, response, props, id, values, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-budgets", erpview.NoticeUpdated))
}
