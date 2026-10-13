package api_get_account

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/get-account: {"account": {...}}, the account
// with the id the body names; 404 when none has it.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	record, ok := erpstore.FindAccount(sandbox, int64(input.Body.Id))
	if !ok {
		return erpapi.Fail(sandbox, props, api.StatusNotFound, "id", "No record has that id.")
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	balance := erpreports.AccountBalance{ByCompany: map[int64]int64{}}
	for _, found := range erpreports.AccountBalances(sandbox, book, book.Today) {
		if found.Account.Id == record.Id {
			balance = found
		}
	}
	shares := balance.ByCompany
	return erpapi.One(sandbox, response, api.StatusOK, "account", erpapi.AccountJSON(sandbox, record, balance.Total, balance.Pending, shares))
}
