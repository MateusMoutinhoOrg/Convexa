package api_list_transactions

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-transactions: {"transactions": [...],
// "total", "page", "pages", "limit"}, the transactions the body's filters
// keep, newest first; limit defaults to 100, at most 1000.
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
	invoice, err := erpapi.Month(sandbox, props, "card-invoice", input.Body.CardInvoice)
	if err != nil {
		return err
	}
	txs, err := erpstore.ListTransactions(sandbox, erpstore.TransactionQuery{
		Search: input.Body.Search, Company: int64(input.Body.Company), Account: int64(input.Body.Account),
		Category: int64(input.Body.Category), ThirdParty: int64(input.Body.ThirdParty), Type: input.Body.Type,
		Status: input.Body.Status, From: from, To: to, DateField: input.Body.DateField, CardInvoice: invoice,
		InstallmentGroup: int64(input.Body.InstallmentGroup), Recurrence: int64(input.Body.Recurrence),
	})
	if err != nil {
		return err
	}
	limit := input.Body.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	pages := (len(txs) + limit - 1) / limit
	if pages < 1 {
		pages = 1
	}
	page := input.Body.Page
	if page < 1 {
		page = 1
	}
	start := (page - 1) * limit
	if start > len(txs) {
		start = len(txs)
	}
	end := start + limit
	if end > len(txs) {
		end = len(txs)
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("transactions", erpapi.TransactionsJSON(sandbox, txs[start:end], erpdate.Today(sandbox)))
	document.AddItemToObject("total", len(txs))
	document.AddItemToObject("page", page)
	document.AddItemToObject("pages", pages)
	document.AddItemToObject("limit", limit)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
