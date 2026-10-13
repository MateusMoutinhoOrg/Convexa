package add_account_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/add-account: a account the form describes
// well is added and the browser sent to the list; anything else answers the
// form again, filled with what was sent, under a 400 with the reason above it.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	reader := erpview.Reader{}
	fields := erpstore.AccountFields{
		Name: input.Body.Name, Kind: input.Body.Kind, Institution: input.Body.Institution, Notes: input.Body.Notes, Status: input.Body.Status,
		CreditLimit:    erpview.Money(sandbox, &reader, "credit-limit", input.Body.CreditLimit),
		ClosingDay:     erpview.Int(sandbox, &reader, "closing-day", input.Body.ClosingDay),
		DueDay:         erpview.Int(sandbox, &reader, "due-day", input.Body.DueDay),
		PaymentAccount: erpview.ParseId(sandbox, input.Body.PaymentAccount),
		OpeningBalance: erpview.Money(sandbox, &reader, "opening-balance", input.Body.OpeningBalance),
		OpeningCompany: erpview.ParseId(sandbox, input.Body.OpeningCompany),
		OpeningDate:    erpview.Date(sandbox, &reader, "opening-date", input.Body.OpeningDate),
	}
	problem := erpview.ProblemOf(sandbox, reader)
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, problem, err = erpstore.AddAccount(sandbox, fields, props.User.Id)
		if err != nil {
			return err
		}
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderAccountForm(sandbox, response, props, 0, fields, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-accounts", erpview.NoticeAdded))
}
