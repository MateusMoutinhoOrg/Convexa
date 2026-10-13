package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
)

// The projected balance is where the cash of the companies is heading: from
// today's balance, every income and expense still to come — the pending
// transactions, the occurrences of the recurrences, the card invoices — day
// by day up to a horizon, a year by default. It is drawn twice: with only
// what is registered, and with the budgets too, counting what is left of each
// budget as still to happen by the end of its period.

// MaxProjectionMonths is the farthest a projection looks ahead.
const MaxProjectionMonths = 60

// ProjectionQuery is what a projection is computed for.
type ProjectionQuery struct {
	// Months is how far ahead it looks: 12 when 0.
	Months int64
	// Company and Account narrow it to one company and one cash account; 0
	// is every one. The budgets belong to no account, so they count only
	// when every account does.
	Company int64
	Account int64
}

// ProjectionPoint is the projected balance at the end of one day.
type ProjectionPoint struct {
	Date        int64
	Balance     int64
	WithBudgets int64
}

// ProjectionMonth is what one month of the projection holds.
type ProjectionMonth struct {
	Month   int64
	Opening int64
	Income  int64
	Expense int64
	// Budgets is what is left of the budgets ending this month, signed.
	Budgets            int64
	Closing            int64
	ClosingWithBudgets int64
}

// Projection is the projected balance of a scope.
type Projection struct {
	Query ProjectionQuery
	Start int64
	End   int64
	// Opening is today's settled balance.
	Opening          int64
	Final            int64
	FinalWithBudgets int64
	// Lowest is the lowest the balance gets, on LowestDate; likewise with
	// the budgets.
	Lowest                int64
	LowestDate            int64
	LowestWithBudgets     int64
	LowestWithBudgetsDate int64
	Income                int64
	Expense               int64
	BudgetTotal           int64
	Months                []ProjectionMonth
	Points                []ProjectionPoint
	// Events are every income and expense to come, the budgets' included.
	Events []Event
	// HasBudgets is false when the projection is narrowed to one account.
	HasBudgets bool
}

// ComputeProjection is the projection query asks for, over book.
func ComputeProjection(sandbox *api.Sandbox, book Book, query ProjectionQuery) Projection {
	if query.Months <= 0 {
		query.Months = 12
	}
	if query.Months > MaxProjectionMonths {
		query.Months = MaxProjectionMonths
	}
	if IsCardAccount(sandbox, book, query.Account) {
		query.Account = 0
	}
	scope := CashScope(sandbox, book, query.Account, query.Company)
	start := book.Today
	end := erpdate.AddMonths(sandbox, start, query.Months)
	projection := Projection{Query: query, Start: start, End: end, HasBudgets: scope.AllCash}
	projection.Opening = Balance(sandbox, book, scope, start)

	ahead := Events(sandbox, book, scope, EventOptions{From: erpdate.AddDays(sandbox, start, 1), To: end, Realized: true})
	expected := Events(sandbox, book, scope, EventOptions{From: start, To: end, Pending: true, Recurrences: true, Invoices: true, Budgets: true})
	events := append(ahead, expected...)
	sandbox.Deps.SortDeps.SliceStable(events, func(i int, j int) bool { return events[i].Date < events[j].Date })
	projection.Events = events

	byDay := map[int64]int64{}
	budgetByDay := map[int64]int64{}
	months := map[int64]*ProjectionMonth{}
	monthOf := func(date int64) *ProjectionMonth {
		month := erpdate.Month(sandbox, date)
		found, ok := months[month]
		if !ok {
			found = &ProjectionMonth{Month: month}
			months[month] = found
		}
		return found
	}
	for _, event := range events {
		month := monthOf(event.Date)
		if event.Source == SourceBudget {
			budgetByDay[event.Date] += event.Amount
			month.Budgets += event.Amount
			projection.BudgetTotal += event.Amount
			continue
		}
		byDay[event.Date] += event.Amount
		if event.Amount >= 0 {
			month.Income += event.Amount
			projection.Income += event.Amount
		} else {
			month.Expense -= event.Amount
			projection.Expense -= event.Amount
		}
	}

	balance, withBudgets := projection.Opening, projection.Opening
	projection.Lowest, projection.LowestDate = balance, start
	projection.LowestWithBudgets, projection.LowestWithBudgetsDate = balance, start
	for date := start; date <= end; date = erpdate.AddDays(sandbox, date, 1) {
		balance += byDay[date]
		withBudgets += byDay[date] + budgetByDay[date]
		projection.Points = append(projection.Points, ProjectionPoint{Date: date, Balance: balance, WithBudgets: withBudgets})
		if balance < projection.Lowest {
			projection.Lowest, projection.LowestDate = balance, date
		}
		if withBudgets < projection.LowestWithBudgets {
			projection.LowestWithBudgets, projection.LowestWithBudgetsDate = withBudgets, date
		}
	}
	projection.Final, projection.FinalWithBudgets = balance, withBudgets

	opening, openingWithBudgets := projection.Opening, projection.Opening
	for month := erpdate.Month(sandbox, start); month <= erpdate.Month(sandbox, end); month = erpdate.AddToMonth(sandbox, month, 1) {
		row := ProjectionMonth{Month: month}
		if found, ok := months[month]; ok {
			row = *found
		}
		row.Opening = opening
		row.Closing = opening + row.Income - row.Expense
		row.ClosingWithBudgets = openingWithBudgets + row.Income - row.Expense + row.Budgets
		opening, openingWithBudgets = row.Closing, row.ClosingWithBudgets
		projection.Months = append(projection.Months, row)
	}
	return projection
}
