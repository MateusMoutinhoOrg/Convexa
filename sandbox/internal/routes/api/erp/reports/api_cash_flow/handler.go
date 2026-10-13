package api_cash_flow

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/cash-flow: the cash flow the body asks for:
// group day, week or month; basis realized or projected.
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
	flow := erpreports.ComputeCashFlow(sandbox, book, erpreports.CashFlowQuery{
		From: from, To: to, Group: input.Body.Group, Basis: input.Body.Basis, Company: int64(input.Body.Company),
		Account: int64(input.Body.Account), Category: int64(input.Body.Category), ThirdParty: int64(input.Body.ThirdParty),
	})
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.CashFlowJSON(sandbox, flow))
}
