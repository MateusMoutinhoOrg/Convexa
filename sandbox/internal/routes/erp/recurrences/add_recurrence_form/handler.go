package add_recurrence_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/add-recurrence: the recurrence is
// written from the form — its occurrences up to the horizon as pending
// transactions — and the browser sent to the list; anything else answers the
// form again under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	id := int64(0)
	values := erpview.RecurrenceForm{
		Company: input.Body.Company, Account: input.Body.Account, Category: input.Body.Category, ThirdParty: input.Body.ThirdParty,
		Type: input.Body.Type, Amount: input.Body.Amount, Description: input.Body.Description, Notes: input.Body.Notes,
		Frequency: input.Body.Frequency, Interval: input.Body.Interval, StartDate: input.Body.StartDate,
		EndDate: input.Body.EndDate, Status: input.Body.Status,
	}
	fields, problem := erpview.RecurrenceFieldsOf(sandbox, values)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, problem, err = erpstore.AddRecurrence(sandbox, fields, props.User.Id)
		if err != nil {
			return err
		}
	}
	if problem.Status == api.StatusNotFound {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-recurrences", erpview.NoticeNotFound))
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderRecurrenceForm(sandbox, response, props, id, values, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-recurrences", erpview.NoticeAdded))
}
