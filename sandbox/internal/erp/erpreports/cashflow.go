package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// The cash flow is the money that came into and left the cash accounts over a
// period, grouped by day, week or month: what was settled — the realized
// basis — or, with the projected basis, what is still expected too, the
// pending transactions, the recurrences and the card invoices included.
// Credit cards hold no cash: a purchase enters the cash flow when its invoice
// is paid, unless the report is narrowed to the card itself.

// The groupings of a cash flow.
const (
	GroupDay   = "day"
	GroupWeek  = "week"
	GroupMonth = "month"
)

// The bases of a cash flow.
const (
	BasisRealized  = "realized"
	BasisProjected = "projected"
)

// CashFlowQuery is what a cash flow is computed for.
type CashFlowQuery struct {
	From  int64
	To    int64
	Group string
	Basis string
	// Company, Account, Category and ThirdParty narrow it; 0 is every one.
	// With a category or a third party, the opening balance is not shown:
	// only the flows of that category or party are.
	Company    int64
	Account    int64
	Category   int64
	ThirdParty int64
}

// CashFlowRow is one period of a cash flow.
type CashFlowRow struct {
	Start    int64
	End      int64
	Inflows  int64
	Outflows int64
	Net      int64
	// Balance is the balance at the end of the period.
	Balance int64
}

// CashFlow is a cash flow and its totals.
type CashFlow struct {
	Query    CashFlowQuery
	Opening  int64
	Inflows  int64
	Outflows int64
	Net      int64
	Closing  int64
	// HasBalance is false when a category or a third party narrows the
	// report, which leaves the balances out.
	HasBalance bool
	Rows       []CashFlowRow
}

// NormalizeCashFlowQuery fills what query leaves blank: this month, grouped
// by day, on the realized basis.
func NormalizeCashFlowQuery(sandbox *api.Sandbox, query CashFlowQuery) CashFlowQuery {
	today := erpdate.Today(sandbox)
	if query.From == 0 {
		query.From = erpdate.MonthStart(sandbox, today)
	}
	if query.To == 0 || query.To < query.From {
		query.To = erpdate.MonthEnd(sandbox, query.From)
	}
	if erpdate.DaysBetween(sandbox, query.From, query.To) > 3660 {
		query.To = erpdate.AddDays(sandbox, query.From, 3660)
	}
	if query.Group != GroupDay && query.Group != GroupWeek && query.Group != GroupMonth {
		query.Group = GroupDay
		if erpdate.DaysBetween(sandbox, query.From, query.To) > 62 {
			query.Group = GroupMonth
		}
	}
	if query.Basis != BasisProjected {
		query.Basis = BasisRealized
	}
	return query
}

// periodStart is the first day of the period of group date lies in: the
// week starts on Monday.
func periodStart(sandbox *api.Sandbox, group string, date int64) int64 {
	switch group {
	case GroupMonth:
		return erpdate.MonthStart(sandbox, date)
	case GroupWeek:
		back := (erpdate.Weekday(sandbox, date) + 6) % 7
		return erpdate.AddDays(sandbox, date, -back)
	}
	return date
}

// periodEnd is the last day of the period of group starting on start.
func periodEnd(sandbox *api.Sandbox, group string, start int64) int64 {
	switch group {
	case GroupMonth:
		return erpdate.MonthEnd(sandbox, start)
	case GroupWeek:
		return erpdate.AddDays(sandbox, start, 6)
	}
	return start
}

// ComputeCashFlow is the cash flow query asks for, over book.
func ComputeCashFlow(sandbox *api.Sandbox, book Book, query CashFlowQuery) CashFlow {
	query = NormalizeCashFlowQuery(sandbox, query)
	scope := CashScope(sandbox, book, query.Account, query.Company)
	flow := CashFlow{Query: query, HasBalance: query.Category == 0 && query.ThirdParty == 0}
	if flow.HasBalance {
		flow.Opening = Balance(sandbox, book, scope, erpdate.AddDays(sandbox, query.From, -1))
	}
	var subtree map[int64]bool
	if query.Category != 0 {
		subtree = erpstore.CategorySubtree(sandbox, book.Categories, query.Category)
	}

	amounts := map[int64]int64{}
	inflows := map[int64]int64{}
	outflows := map[int64]int64{}
	add := func(date int64, amount int64) {
		if date < query.From || date > query.To || amount == 0 {
			return
		}
		start := periodStart(sandbox, query.Group, date)
		amounts[start] += amount
		if amount > 0 {
			inflows[start] += amount
		} else {
			outflows[start] -= amount
		}
	}
	events := Events(sandbox, book, scope, EventOptions{
		From: query.From, To: query.To, Realized: true, Pending: query.Basis == BasisProjected,
		Recurrences: query.Basis == BasisProjected, Invoices: query.Basis == BasisProjected && query.Category == 0 && query.ThirdParty == 0,
	})
	for _, event := range events {
		if subtree != nil && !subtree[event.Category] {
			continue
		}
		if query.ThirdParty != 0 && event.ThirdParty != query.ThirdParty {
			continue
		}
		add(event.Date, event.Amount)
	}

	balance := flow.Opening
	for start := periodStart(sandbox, query.Group, query.From); start <= query.To; start = erpdate.AddDays(sandbox, periodEnd(sandbox, query.Group, start), 1) {
		row := CashFlowRow{Start: start, End: periodEnd(sandbox, query.Group, start), Inflows: inflows[start], Outflows: outflows[start], Net: amounts[start]}
		if row.Start < query.From {
			row.Start = query.From
		}
		if row.End > query.To {
			row.End = query.To
		}
		balance += row.Net
		row.Balance = balance
		flow.Inflows += row.Inflows
		flow.Outflows += row.Outflows
		flow.Rows = append(flow.Rows, row)
	}
	flow.Net = flow.Inflows - flow.Outflows
	flow.Closing = flow.Opening + flow.Net
	return flow
}
