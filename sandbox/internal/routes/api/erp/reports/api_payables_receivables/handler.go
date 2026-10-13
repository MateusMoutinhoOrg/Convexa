package api_payables_receivables

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/payables-receivables: what is still to pay
// and to receive, as the body narrows it: kind payable or receivable, due from
// and up to two dates, overdue only.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	from, err := erpapi.Date(sandbox, props, "from", input.Body.From)
	if err != nil {
		return err
	}
	to, err := erpapi.Date(sandbox, props, "to", input.Body.To)
	if err != nil {
		return err
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	list := erpreports.ComputeDue(sandbox, book, erpreports.DueQuery{
		Kind: input.Body.Kind, Company: int64(input.Body.Company), Account: int64(input.Body.Account),
		ThirdParty: int64(input.Body.ThirdParty), From: from, To: to, OverdueOnly: input.Body.Overdue,
	})
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.DueJSON(sandbox, list))
}
