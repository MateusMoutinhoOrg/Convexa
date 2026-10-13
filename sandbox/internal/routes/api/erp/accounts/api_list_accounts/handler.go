package api_list_accounts

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-accounts: {"accounts": [...]}, every
// account the body's filters keep.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	records, err := erpstore.Accounts(sandbox, erpstore.AccountQuery{Search: input.Body.Search, Kind: input.Body.Kind, Status: input.Body.Status, Cards: input.Body.Cards})
	if err != nil {
		return err
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	balances := map[int64]erpreports.AccountBalance{}
	for _, balance := range erpreports.AccountBalances(sandbox, book, book.Today) {
		balances[balance.Account.Id] = balance
	}
	list := erpapi.Array(sandbox)
	for _, record := range records {
		balance := balances[record.Id]
		shares := balance.ByCompany
		if shares == nil {
			shares = map[int64]int64{}
		}
		list.AddItemToArray(erpapi.AccountJSON(sandbox, record, balance.Total, balance.Pending, shares))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("accounts", list)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
