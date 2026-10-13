package add_transfer_page

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers GET /admin/manage/add-transfer: the form that adds a
// transfer, its accounts, company, amount and invoice taken from the query when
// given — a card invoice's pay button fills them — dated today, done.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	today := erpdate.ISO(sandbox, erpdate.Today(sandbox))
	values := erpview.TransferForm{
		FromAccount: input.FromAccount, ToAccount: input.ToAccount, FromCompany: input.Company, Amount: input.Amount,
		CardInvoice: input.CardInvoice, Date: today, Settled: "on", Next: input.Next,
	}
	return erpview.RenderTransferForm(sandbox, response, props, 0, values, erpstore.Problem{})
}
