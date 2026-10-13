package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// An Event is money that moved, or is expected to move, in or out of a
// scope on one day: a settled transaction, a pending one, an occurrence of a
// recurrence not written yet, what is left to pay of a card invoice, or what
// is left of a budget. The cash flow, the projected balance and the payables
// are all sums of events.

// The sources of an event.
const (
	SourceSettled    = "settled"
	SourcePending    = "pending"
	SourceRecurrence = "recurrence"
	SourceInvoice    = "invoice"
	SourceBudget     = "budget"
)

// Event is one movement of money, done or expected.
type Event struct {
	Date int64
	// Amount is signed cents: positive in, negative out.
	Amount     int64
	Source     string
	Type       string
	Company    int64
	Account    int64
	Category   int64
	ThirdParty int64
	// Description is the transaction's, the recurrence's, or empty for an
	// invoice and a budget, which a page words from Card, Month and Budget.
	Description string
	// Overdue is true for a pending one that was due before today, and was
	// moved to today.
	Overdue bool
	// TxId is the transaction's id, 0 for an event no transaction holds.
	TxId int64
	// Recurrence is the id of the recurrence of an occurrence not written yet.
	Recurrence int64
	// Card and Month name the invoice of an invoice event.
	Card  int64
	Month int64
	// Budget is the id of the budget of a budget event.
	Budget int64
}

// EventOptions is what Events lists.
type EventOptions struct {
	// From and To bound the dates of the events, both included.
	From int64
	To   int64
	// Realized lists the transactions settled in the period.
	Realized bool
	// Pending lists the pending transactions expected in the period, the
	// overdue ones on today.
	Pending bool
	// Recurrences lists the occurrences of the recurrences not written as
	// transactions yet.
	Recurrences bool
	// Invoices lists what is left to pay of the card invoices due in the
	// period, on the accounts that pay them.
	Invoices bool
	// Budgets lists what is left of each budget, on the last day of its
	// period: what it plans beyond what the transactions and recurrences
	// already hold.
	Budgets bool
}

// eventOf is the event of tx on date, from source.
func eventOf(sandbox *api.Sandbox, tx erp_db.TransactionRecord, date int64, source string) Event {
	return Event{
		Date: date, Amount: tx.Amount, Source: source, Type: tx.Type, Company: tx.Company, Account: tx.Account,
		Category: tx.Category, ThirdParty: tx.ThirdParty, Description: tx.Description, TxId: tx.Id,
	}
}

// virtualOccurrences are the occurrences of the active recurrences of book
// that no transaction holds yet, falling on or before to.
func virtualOccurrences(sandbox *api.Sandbox, book Book, to int64) []Event {
	events := []Event{}
	for _, recurrence := range book.Recurrences {
		if recurrence.Status != erpstore.RecurrenceActive {
			continue
		}
		for _, date := range erpledger.Occurrences(sandbox, recurrence.StartDate, recurrence.Frequency, recurrence.Interval, recurrence.EndDate, recurrence.GeneratedUntil, to) {
			amount := recurrence.Amount
			if recurrence.Type == erpledger.TypeExpense {
				amount = -amount
			}
			events = append(events, Event{
				Date: date, Amount: amount, Source: SourceRecurrence, Type: recurrence.Type, Company: recurrence.Company,
				Account: recurrence.Account, Category: recurrence.Category, ThirdParty: recurrence.ThirdParty,
				Description: recurrence.Description, Recurrence: recurrence.Id,
			})
		}
	}
	return events
}

// onOrAfterToday is date, or today when date has passed: a pending event
// that is overdue is still expected, today at the earliest.
func onOrAfterToday(sandbox *api.Sandbox, book Book, date int64) (int64, bool) {
	if date < book.Today {
		return book.Today, true
	}
	return date, false
}

// Events is every event of scope that options asks for, by date.
func Events(sandbox *api.Sandbox, book Book, scope Scope, options EventOptions) []Event {
	byId := counterparts(sandbox, book)
	events := []Event{}
	within := func(date int64) bool { return date >= options.From && date <= options.To }

	for _, tx := range book.Transactions {
		if !In(sandbox, scope, tx.Account, tx.Company) || Internal(sandbox, scope, byId, tx) {
			continue
		}
		if tx.Status == erpledger.StatusSettled {
			if options.Realized && within(tx.SettledDate) {
				events = append(events, eventOf(sandbox, tx, tx.SettledDate, SourceSettled))
			}
			continue
		}
		if options.Pending {
			date, overdue := onOrAfterToday(sandbox, book, erpledger.ExpectedDate(sandbox, tx))
			if within(date) {
				event := eventOf(sandbox, tx, date, SourcePending)
				event.Overdue = overdue
				events = append(events, event)
			}
		}
	}

	var virtual []Event
	if options.Recurrences || options.Invoices || options.Budgets {
		virtual = virtualOccurrences(sandbox, book, options.To)
	}
	if options.Recurrences {
		for _, event := range virtual {
			if !In(sandbox, scope, event.Account, event.Company) {
				continue
			}
			if IsCardAccount(sandbox, book, event.Account) {
				event.Source = SourceSettled
			}
			date, overdue := onOrAfterToday(sandbox, book, event.Date)
			if within(date) {
				event.Date, event.Overdue = date, overdue
				events = append(events, event)
			}
		}
	}
	if options.Invoices {
		events = append(events, invoiceEvents(sandbox, book, scope, virtual, options)...)
	}
	if options.Budgets && scope.AllCash {
		events = append(events, budgetEvents(sandbox, book, scope, virtual, options)...)
	}
	sandbox.Deps.SortDeps.SliceStable(events, func(i int, j int) bool { return events[i].Date < events[j].Date })
	return events
}

// invoiceEvents are what is left to pay of every card invoice due in the
// period of options — the occurrences of the recurrences charged to the
// cards, not written yet, counted in — as one event per company that owes
// it, on the account that pays the card. What a pending payment already
// covers is left out: that payment is an event of its own.
func invoiceEvents(sandbox *api.Sandbox, book Book, scope Scope, virtual []Event, options EventOptions) []Event {
	events := []Event{}
	for _, card := range Cards(sandbox, book) {
		payer := card.PaymentAccount
		if payer == 0 {
			if !scope.AllCash {
				continue
			}
		} else if !scope.Accounts[payer] {
			continue
		}
		type key struct{ month, company int64 }
		left := map[key]int64{}
		keys := []key{}
		touch := func(k key, amount int64) {
			if _, ok := left[k]; !ok {
				keys = append(keys, k)
			}
			left[k] += amount
		}
		for _, tx := range book.Transactions {
			if tx.Account != card.Id || tx.CardInvoice == 0 {
				continue
			}
			touch(key{tx.CardInvoice, tx.Company}, -tx.Amount)
		}
		for _, event := range virtual {
			if event.Account == card.Id {
				touch(key{erpledger.InvoiceFor(sandbox, card, event.Date), event.Company}, -event.Amount)
			}
		}
		for _, k := range keys {
			if left[k] <= 0 || (scope.Company != 0 && k.company != scope.Company) {
				continue
			}
			date, overdue := onOrAfterToday(sandbox, book, erpledger.InvoiceDue(sandbox, card, k.month))
			if date < options.From || date > options.To {
				continue
			}
			events = append(events, Event{
				Date: date, Amount: -left[k], Source: SourceInvoice, Type: erpledger.TypeExpense, Company: k.company,
				Account: payer, Overdue: overdue, Card: card.Id, Month: k.month,
			})
		}
	}
	return events
}

// budgetEvents are what is left of every budget of the company in scope that
// ends on or after today: the budget less what the transactions — on any
// account, the cards included — and the occurrences not written yet of its
// category and the ones under it hold in its period, on the last day of the
// period. A budget already reached leaves nothing.
func budgetEvents(sandbox *api.Sandbox, book Book, scope Scope, virtual []Event, options EventOptions) []Event {
	events := []Event{}
	for _, budget := range book.Budgets {
		if budget.PeriodEnd < book.Today || (scope.Company != 0 && budget.Company != scope.Company) {
			continue
		}
		date := budget.PeriodEnd
		if date < options.From || date > options.To {
			continue
		}
		category := book.CategoryById[budget.Category]
		committed := Committed(sandbox, book, budget, virtual)
		left := budget.Amount - committed
		if left <= 0 {
			continue
		}
		amount := left
		if category.Type == erpledger.TypeExpense {
			amount = -left
		}
		events = append(events, Event{
			Date: date, Amount: amount, Source: SourceBudget, Type: category.Type, Company: budget.Company,
			Category: budget.Category, Budget: budget.Id,
		})
	}
	return events
}

// Committed is how much of budget the transactions of its company, its
// category or one under it, dated in its period, hold — settled or not —
// plus the occurrences of virtual in it, as a positive amount.
func Committed(sandbox *api.Sandbox, book Book, budget erp_db.BudgetRecord, virtual []Event) int64 {
	realized, pending := BudgetActual(sandbox, book, budget)
	committed := realized + pending
	subtree := erpstore.CategorySubtree(sandbox, book.Categories, budget.Category)
	category := book.CategoryById[budget.Category]
	for _, event := range virtual {
		if event.Company != budget.Company || !subtree[event.Category] || event.Date < budget.PeriodStart || event.Date > budget.PeriodEnd {
			continue
		}
		if category.Type == erpledger.TypeExpense {
			committed -= event.Amount
		} else {
			committed += event.Amount
		}
	}
	return committed
}

// BudgetActual is how much of budget the transactions of its company, its
// category or one under it, dated in its period, hold: the settled ones and
// the pending ones, as positive amounts.
func BudgetActual(sandbox *api.Sandbox, book Book, budget erp_db.BudgetRecord) (int64, int64) {
	subtree := erpstore.CategorySubtree(sandbox, book.Categories, budget.Category)
	category := book.CategoryById[budget.Category]
	var realized, pending int64
	for _, tx := range book.Transactions {
		if tx.Company != budget.Company || !subtree[tx.Category] || tx.Date < budget.PeriodStart || tx.Date > budget.PeriodEnd {
			continue
		}
		if tx.Type != erpledger.TypeIncome && tx.Type != erpledger.TypeExpense {
			continue
		}
		amount := tx.Amount
		if category.Type == erpledger.TypeExpense {
			amount = -amount
		}
		if tx.Status == erpledger.StatusSettled {
			realized += amount
		} else {
			pending += amount
		}
	}
	return realized, pending
}
