package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// The payables and receivables are what the companies still have to pay and
// to receive: the pending transactions of the cash accounts, and what is left
// to pay of the card invoices — one per company that owes it. A pending
// transfer is no payable, unless it pays a card invoice.

// The kinds of a due item.
const (
	KindPayable    = "payable"
	KindReceivable = "receivable"
)

// DueQuery narrows the payables and receivables.
type DueQuery struct {
	// Kind keeps payables or receivables, "" both.
	Kind       string
	Company    int64
	Account    int64
	ThirdParty int64
	// From and To bound the due dates, both included; 0 is open. Overdue
	// items are kept whatever From says.
	From int64
	To   int64
	// OverdueOnly keeps the items due before today.
	OverdueOnly bool
}

// DueItem is one payable or receivable.
type DueItem struct {
	Event
	Kind string
	// DueDate is when it is due; Event.Date is when it is expected, today
	// for an overdue one.
	DueDate int64
	// Amount is positive: Kind says which way it goes.
	Amount int64
}

// DueList is the payables and receivables a DueQuery keeps, by due date, and
// their totals.
type DueList struct {
	Query             DueQuery
	Items             []DueItem
	Payable           int64
	Receivable        int64
	OverduePayable    int64
	OverdueReceivable int64
}

// ComputeDue is the payables and receivables query keeps, over book.
func ComputeDue(sandbox *api.Sandbox, book Book, query DueQuery) DueList {
	list := DueList{Query: query}
	scope := CashScope(sandbox, book, query.Account, query.Company)
	to := query.To
	if to == 0 {
		to = 99991231
	}
	events := Events(sandbox, book, scope, EventOptions{From: 0, To: to, Pending: true, Invoices: true})
	byId := counterparts(sandbox, book)
	for _, event := range events {
		item := DueItem{Event: event, Kind: KindPayable, Amount: -event.Amount, DueDate: event.Date}
		if event.Amount > 0 {
			item.Kind, item.Amount = KindReceivable, event.Amount
		}
		if event.TxId != 0 {
			tx := byId[event.TxId]
			item.DueDate = erpledger.ExpectedDate(sandbox, tx)
			if erpledger.IsTransfer(sandbox, tx.Type) && !paysInvoice(sandbox, book, byId, tx) {
				continue
			}
		}
		if event.Source == SourceInvoice {
			item.DueDate = erpledger.InvoiceDue(sandbox, book.AccountById[event.Card], event.Month)
		}
		if query.Kind != "" && item.Kind != query.Kind {
			continue
		}
		if query.ThirdParty != 0 && item.ThirdParty != query.ThirdParty {
			continue
		}
		if query.OverdueOnly && !item.Overdue {
			continue
		}
		if !item.Overdue && query.From != 0 && item.DueDate < query.From {
			continue
		}
		list.Items = append(list.Items, item)
		if item.Kind == KindPayable {
			list.Payable += item.Amount
			if item.Overdue {
				list.OverduePayable += item.Amount
			}
		} else {
			list.Receivable += item.Amount
			if item.Overdue {
				list.OverdueReceivable += item.Amount
			}
		}
	}
	sandbox.Deps.SortDeps.SliceStable(list.Items, func(i int, j int) bool { return list.Items[i].DueDate < list.Items[j].DueDate })
	return list
}

// paysInvoice tells whether tx, a transfer leg, is the payment of a card
// invoice: its other leg, found in byId, reaches a card.
func paysInvoice(sandbox *api.Sandbox, book Book, byId map[int64]erp_db.TransactionRecord, tx erp_db.TransactionRecord) bool {
	if tx.Type != erpledger.TypeTransferOut {
		return false
	}
	return IsCardAccount(sandbox, book, byId[tx.Counterpart].Account)
}
