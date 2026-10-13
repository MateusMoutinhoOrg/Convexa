package list_transactions_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/erp/list-transactions: the transactions the filter
// form keeps — this month's, when it names no period — newest first, a page at
// a time, with their totals.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	filters := erpview.TransactionFilters{
		Search: input.Search, Company: input.Company, Account: input.Account, Category: input.Category,
		ThirdParty: input.ThirdParty, Type: input.Type, Status: input.Status, From: input.From, To: input.To,
		DateField: input.DateField, CardInvoice: input.CardInvoice, InstallmentGroup: input.InstallmentGroup,
		Recurrence: input.Recurrence, Page: erpview.ParseInt(sandbox, input.Page),
	}
	return erpview.RenderTransactionsPage(sandbox, response, props, filters, input.Notice)
}
