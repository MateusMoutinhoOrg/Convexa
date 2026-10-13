package api_account_balances

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/account-balances: the balance of every
// account at the body's date — today when blank — whole, pending and by
// company, with the totals of the cash and of the card debt.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	date, err := erpapi.Date(sandbox, props, "date", input.Body.Date)
	if err != nil {
		return err
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	if date == 0 {
		date = book.Today
	}
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.BalancesJSON(sandbox, erpreports.AccountBalances(sandbox, book, date)))
}
