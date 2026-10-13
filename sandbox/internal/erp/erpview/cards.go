package erpview

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// CardsPage is what erp/cards.html is rendered with.
type CardsPage struct {
	Cards []CardView
}

// CardView is one credit card as the pages show it.
type CardView struct {
	Summary erpreports.CardSummary
	// PaymentAccount names the account that pays its invoices.
	PaymentAccount string
	// Used is how much of the limit the debt takes, in percent, at most 100.
	Used int64
}

// RenderCardsPage answers the list of the credit cards, with the notice
// code notice above it.
func RenderCardsPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	body := CardsPage{}
	for _, summary := range erpreports.ComputeCardSummaries(sandbox, book) {
		view := CardView{Summary: summary, PaymentAccount: book.AccountById[summary.Card.PaymentAccount].Name}
		if summary.Card.CreditLimit > 0 {
			view.Used = summary.Debt * 100 / summary.Card.CreditLimit
			if view.Used > 100 {
				view.Used = 100
			}
			if view.Used < 0 {
				view.Used = 0
			}
		}
		body.Cards = append(body.Cards, view)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/cards.html", "Credit cards", "cards", NoticeOf(sandbox, notice), body)
}

// InvoicesPage is what erp/card_invoices.html is rendered with.
type InvoicesPage struct {
	Card     erp_db.AccountRecord
	Invoices []erpledger.Invoice
	Summary  erpreports.CardSummary
}

// RenderInvoicesPage answers the invoices of the card with id card, newest
// first; the list of the cards, when it is none.
func RenderInvoicesPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, card int64, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	account, ok := book.AccountById[card]
	if !ok || !erpledger.IsCard(sandbox, account) {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, WithNotice(sandbox, "/admin/erp/list-credit-cards", NoticeNotFound))
	}
	invoices := erpledger.Invoices(sandbox, account, book.Transactions, book.Today)
	for i, j := 0, len(invoices)-1; i < j; i, j = i+1, j-1 {
		invoices[i], invoices[j] = invoices[j], invoices[i]
	}
	body := InvoicesPage{Card: account, Invoices: invoices}
	for _, summary := range erpreports.ComputeCardSummaries(sandbox, book) {
		if summary.Card.Id == card {
			body.Summary = summary
		}
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/card_invoices.html", account.Name, "cards", NoticeOf(sandbox, notice), body)
}

// ShareView is what one company owes on an invoice.
type ShareView struct {
	Company string
	Share   erpledger.InvoiceShare
}

// InvoicePage is what erp/card_invoice.html is rendered with.
type InvoicePage struct {
	Card    erp_db.AccountRecord
	Invoice erpledger.Invoice
	// MonthISO spells the invoice's month for the pay form, 2026-10.
	MonthISO string
	// ToPay is what paying it now pays: what is left of it, less what a
	// payment already scheduled covers.
	ToPay  int64
	Rows   []TxRow
	Shares []ShareView
	Payers []Option
	// PayDate fills the date of the pay form: the due date, or today when
	// it has passed.
	PayDate string
	// Prev and Next are the months before and after, for the arrows.
	Prev string
	Next string
	// Self is this very page, which a form returns to.
	Self string
}

// RenderInvoicePage answers the invoice of month (2026-10, "" for the one
// open today) of the card with id card.
func RenderInvoicePage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, card int64, month string, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	account, ok := book.AccountById[card]
	if !ok || !erpledger.IsCard(sandbox, account) {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, WithNotice(sandbox, "/admin/erp/list-credit-cards", NoticeNotFound))
	}
	value, ok := erpdate.ParseMonth(sandbox, month)
	if !ok {
		value = erpledger.InvoiceFor(sandbox, account, book.Today)
	}
	invoice := erpledger.FindInvoice(sandbox, account, book.Transactions, value, book.Today)
	txs := erpstore.FilterTransactions(sandbox, book.Transactions, book.Categories, erpstore.TransactionQuery{Account: card, CardInvoice: value})
	cardId := sandbox.Deps.StringsDeps.FormatInt(card, 10)
	body := InvoicePage{
		Card: account, Invoice: invoice, MonthISO: erpdate.ISOMonth(sandbox, value), Rows: Rows(sandbox, book, txs),
		Payers:  AccountOptions(sandbox, book.Accounts, account.PaymentAccount, false),
		PayDate: erpdate.ISO(sandbox, invoice.Due),
		Prev:    erpdate.ISOMonth(sandbox, erpdate.AddToMonth(sandbox, value, -1)),
		Next:    erpdate.ISOMonth(sandbox, erpdate.AddToMonth(sandbox, value, 1)),
		Self:    Href(sandbox, "/admin/erp/card-invoice/"+cardId, "month", erpdate.ISOMonth(sandbox, value)),
	}
	if invoice.Due < book.Today {
		body.PayDate = erpdate.ISO(sandbox, book.Today)
	}
	body.ToPay = invoice.Remaining - invoice.Scheduled
	for _, share := range erpledger.InvoiceShares(sandbox, account, book.Transactions, value) {
		body.Shares = append(body.Shares, ShareView{Company: book.CompanyById[share.Company].Name, Share: share})
	}
	title := erpdate.FormatMonth(sandbox, erppage.Language(sandbox, props), value)
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/card_invoice.html", account.Name+" · "+title, "cards", NoticeOf(sandbox, notice), body)
}
