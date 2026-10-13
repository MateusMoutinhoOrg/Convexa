package add_transaction_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/manage/add-transaction: the form that adds a
// transaction, its type, company, account and invoice taken from the query
// when given, dated today.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	today := erpdate.ISO(sandbox, erpdate.Today(sandbox))
	values := erpview.TransactionForm{
		Type: input.Type, Company: input.Company, Account: input.Account, CardInvoice: input.CardInvoice,
		Date: today, DueDate: today, Next: input.Next,
	}
	return erpview.RenderTransactionForm(sandbox, response, props, 0, values, erpstore.Problem{})
}
