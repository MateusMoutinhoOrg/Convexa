package erpledger

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
)

// The ledger is the arithmetic of Convexa, done over records already read:
// it reads and writes nothing, so every report computes from one listing of
// the transactions, and every rule below is spelled once.
//
// A transaction's amount is signed cents: an inflow is positive, an outflow
// negative. An account's balance is the sum of the amounts of its
// transactions; a company's share of an account is the same sum over the
// transactions of that company alone. A balance at a date counts what was
// settled — paid or received — by then.

// The kinds of account.
const (
	KindChecking   = "checking"
	KindSavings    = "savings"
	KindInvestment = "investment"
	KindWallet     = "wallet"
	KindCreditCard = "credit-card"
	KindOther      = "other"
)

// The types of transaction. Income and expense are what a category
// classifies; the two transfer legs move money between accounts — or
// companies — and an adjustment sets a balance right, an opening balance
// above all.
const (
	TypeIncome      = "income"
	TypeExpense     = "expense"
	TypeTransferIn  = "transfer-in"
	TypeTransferOut = "transfer-out"
	TypeAdjustment  = "adjustment"
)

// The statuses of a transaction.
const (
	StatusPending = "pending"
	StatusSettled = "settled"
)

// The statuses of a card invoice.
const (
	InvoiceOpen    = "open"
	InvoiceClosed  = "closed"
	InvoicePartial = "partial"
	InvoicePaid    = "paid"
	InvoiceOverdue = "overdue"
)

// Kinds is every kind of account, in the order the forms offer them.
func Kinds(sandbox *api.Sandbox) []string {
	return []string{KindChecking, KindSavings, KindInvestment, KindWallet, KindCreditCard, KindOther}
}

// IsCard tells whether account is a credit card.
func IsCard(sandbox *api.Sandbox, account erp_db.AccountRecord) bool {
	return account.Kind == KindCreditCard
}

// IsTransfer tells whether a transaction of type kind is a transfer leg.
func IsTransfer(sandbox *api.Sandbox, kind string) bool {
	return kind == TypeTransferIn || kind == TypeTransferOut
}

// SettledBy tells whether tx was settled on or before date; any settled one
// when date is 0.
func SettledBy(sandbox *api.Sandbox, tx erp_db.TransactionRecord, date int64) bool {
	return tx.Status == StatusSettled && (date == 0 || tx.SettledDate <= date)
}

// ExpectedDate is when tx moves money: its settled date once settled, its
// due date — or its date, with none — while pending.
func ExpectedDate(sandbox *api.Sandbox, tx erp_db.TransactionRecord) int64 {
	if tx.Status == StatusSettled && tx.SettledDate != 0 {
		return tx.SettledDate
	}
	if tx.DueDate != 0 {
		return tx.DueDate
	}
	return tx.Date
}

// Balance is the settled balance at date (0 for "ever") of the account with
// id account (0 for every account) owned by the company with id company (0
// for every company), over txs.
func Balance(sandbox *api.Sandbox, txs []erp_db.TransactionRecord, account int64, company int64, date int64) int64 {
	total := int64(0)
	for _, tx := range txs {
		if account != 0 && tx.Account != account {
			continue
		}
		if company != 0 && tx.Company != company {
			continue
		}
		if SettledBy(sandbox, tx, date) {
			total += tx.Amount
		}
	}
	return total
}

// Pending is the sum of the pending amounts of the account with id account
// (0 for every account) and the company with id company (0 for every
// company), expected on or before date (0 for any date).
func Pending(sandbox *api.Sandbox, txs []erp_db.TransactionRecord, account int64, company int64, date int64) int64 {
	total := int64(0)
	for _, tx := range txs {
		if tx.Status != StatusPending {
			continue
		}
		if account != 0 && tx.Account != account {
			continue
		}
		if company != 0 && tx.Company != company {
			continue
		}
		if date == 0 || ExpectedDate(sandbox, tx) <= date {
			total += tx.Amount
		}
	}
	return total
}

// Invoice is one monthly bill of a credit card.
type Invoice struct {
	// Month is the month the invoice closes in, YYYYMM.
	Month int64
	// Closing is the last day a purchase lands on this invoice.
	Closing int64
	// Due is the day the invoice must be paid by.
	Due int64
	// Charges is what the invoice bills: its purchases less its refunds.
	Charges int64
	// Paid is what settled payments covered of it.
	Paid int64
	// Scheduled is what pending payments will cover of it.
	Scheduled int64
	// Remaining is what is left to pay: Charges less Paid.
	Remaining int64
	Status    string
	// Count is how many purchases and refunds the invoice holds.
	Count int
}

// InvoiceClosing is the day the invoice of month (YYYYMM) of card closes.
func InvoiceClosing(sandbox *api.Sandbox, card erp_db.AccountRecord, month int64) int64 {
	return erpdate.DayOfMonth(sandbox, month, card.ClosingDay)
}

// InvoiceDue is the day the invoice of month (YYYYMM) of card is due: its due
// day in the same month when that comes after the closing day, else in the
// month after.
func InvoiceDue(sandbox *api.Sandbox, card erp_db.AccountRecord, month int64) int64 {
	if card.DueDay > card.ClosingDay {
		return erpdate.DayOfMonth(sandbox, month, card.DueDay)
	}
	return erpdate.DayOfMonth(sandbox, erpdate.AddToMonth(sandbox, month, 1), card.DueDay)
}

// InvoiceFor is the month (YYYYMM) of the invoice a purchase of card on date
// lands on: the one closing that month when it was made on or before the
// closing day, else the next.
func InvoiceFor(sandbox *api.Sandbox, card erp_db.AccountRecord, date int64) int64 {
	month := erpdate.Month(sandbox, date)
	if date <= InvoiceClosing(sandbox, card, month) {
		return month
	}
	return erpdate.AddToMonth(sandbox, month, 1)
}

// IsCharge tells whether tx, a transaction of a card, is billed on an
// invoice — a purchase or a refund — rather than paying one.
func IsCharge(sandbox *api.Sandbox, tx erp_db.TransactionRecord) bool {
	return tx.Type != TypeTransferIn
}

// Invoices is every invoice of card that holds a purchase, a refund or a
// payment in txs, oldest first, as they stand today.
func Invoices(sandbox *api.Sandbox, card erp_db.AccountRecord, txs []erp_db.TransactionRecord, today int64) []Invoice {
	byMonth := map[int64]*Invoice{}
	months := []int64{}
	for _, tx := range txs {
		if tx.Account != card.Id || tx.CardInvoice == 0 {
			continue
		}
		invoice, ok := byMonth[tx.CardInvoice]
		if !ok {
			invoice = &Invoice{Month: tx.CardInvoice}
			byMonth[tx.CardInvoice] = invoice
			months = append(months, tx.CardInvoice)
		}
		if IsCharge(sandbox, tx) {
			invoice.Charges -= tx.Amount
			invoice.Count++
		} else if tx.Status == StatusSettled {
			invoice.Paid += tx.Amount
		} else {
			invoice.Scheduled += tx.Amount
		}
	}
	sandbox.Deps.SortDeps.Slice(months, func(i int, j int) bool { return months[i] < months[j] })
	invoices := []Invoice{}
	for _, month := range months {
		invoice := *byMonth[month]
		invoice.Closing = InvoiceClosing(sandbox, card, month)
		invoice.Due = InvoiceDue(sandbox, card, month)
		invoice.Remaining = invoice.Charges - invoice.Paid
		invoice.Status = invoiceStatus(sandbox, invoice, today)
		invoices = append(invoices, invoice)
	}
	return invoices
}

// FindInvoice is the invoice of month of card over txs, as it stands today;
// an empty one when nothing landed on it yet.
func FindInvoice(sandbox *api.Sandbox, card erp_db.AccountRecord, txs []erp_db.TransactionRecord, month int64, today int64) Invoice {
	for _, invoice := range Invoices(sandbox, card, txs, today) {
		if invoice.Month == month {
			return invoice
		}
	}
	invoice := Invoice{Month: month, Closing: InvoiceClosing(sandbox, card, month), Due: InvoiceDue(sandbox, card, month)}
	invoice.Status = invoiceStatus(sandbox, invoice, today)
	return invoice
}

// invoiceStatus is where invoice stands today.
func invoiceStatus(sandbox *api.Sandbox, invoice Invoice, today int64) string {
	if invoice.Charges > 0 && invoice.Remaining <= 0 {
		return InvoicePaid
	}
	if today <= invoice.Closing {
		return InvoiceOpen
	}
	if invoice.Remaining <= 0 {
		return InvoicePaid
	}
	if today > invoice.Due {
		return InvoiceOverdue
	}
	if invoice.Paid > 0 {
		return InvoicePartial
	}
	return InvoiceClosed
}

// OldestUnpaidInvoice is the month of the oldest invoice of card that has
// closed by date and still has something to pay, or 0 when every closed one
// is paid.
func OldestUnpaidInvoice(sandbox *api.Sandbox, card erp_db.AccountRecord, txs []erp_db.TransactionRecord, date int64) int64 {
	for _, invoice := range Invoices(sandbox, card, txs, date) {
		if invoice.Closing <= date && invoice.Remaining-invoice.Scheduled > 0 {
			return invoice.Month
		}
	}
	return 0
}

// Occurrences are the dates a schedule starting on start, repeating every
// interval weeks, months or years (frequency "weekly", "monthly" or
// "yearly") and ending on end (0 for never), falls on after from and on or
// before to. Each date is counted from start rather than from the one
// before, so a schedule on the 31st falls on the last day of a shorter month
// and returns to the 31st after it.
func Occurrences(sandbox *api.Sandbox, start int64, frequency string, interval int64, end int64, from int64, to int64) []int64 {
	dates := []int64{}
	if interval < 1 {
		interval = 1
	}
	for step := int64(0); step < 100000; step++ {
		var date int64
		switch frequency {
		case "weekly":
			date = erpdate.AddDays(sandbox, start, 7*interval*step)
		case "yearly":
			date = erpdate.AddMonths(sandbox, start, 12*interval*step)
		default:
			date = erpdate.AddMonths(sandbox, start, interval*step)
		}
		if date > to || (end != 0 && date > end) {
			break
		}
		if date > from {
			dates = append(dates, date)
		}
	}
	return dates
}

// InvoiceShare is what one company owes on one invoice of a credit card.
type InvoiceShare struct {
	Company int64
	Charges int64
	// Paid counts the payments settled and the ones scheduled, so a payment
	// already on its way is not paid twice.
	Paid      int64
	Remaining int64
}

// InvoiceShares is what each company owes on the invoice of month of card,
// over txs, by company id.
func InvoiceShares(sandbox *api.Sandbox, card erp_db.AccountRecord, txs []erp_db.TransactionRecord, month int64) []InvoiceShare {
	byCompany := map[int64]*InvoiceShare{}
	companies := []int64{}
	for _, tx := range txs {
		if tx.Account != card.Id || tx.CardInvoice != month {
			continue
		}
		share, ok := byCompany[tx.Company]
		if !ok {
			share = &InvoiceShare{Company: tx.Company}
			byCompany[tx.Company] = share
			companies = append(companies, tx.Company)
		}
		if IsCharge(sandbox, tx) {
			share.Charges -= tx.Amount
		} else {
			share.Paid += tx.Amount
		}
	}
	sandbox.Deps.SortDeps.Slice(companies, func(i int, j int) bool { return companies[i] < companies[j] })
	shares := []InvoiceShare{}
	for _, company := range companies {
		share := *byCompany[company]
		share.Remaining = share.Charges - share.Paid
		shares = append(shares, share)
	}
	return shares
}
