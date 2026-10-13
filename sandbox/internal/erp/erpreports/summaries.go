package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// CategoryTotal is what one category — 0 for the transactions with none —
// adds up to over a period, settled and pending, as positive amounts.
type CategoryTotal struct {
	Category int64
	Type     string
	Realized int64
	Pending  int64
}

// BudgetLine is one budget set against what happened in its period.
type BudgetLine struct {
	Budget   erp_db.BudgetRecord
	Type     string
	Realized int64
	Pending  int64
	// Committed is Realized plus Pending; Remaining is the budget less it,
	// negative when the budget was exceeded.
	Committed int64
	Remaining int64
	// Percent is how much of the budget is committed, 0 to any number.
	Percent int64
}

// CompanyReport is what one company — or every one — did over a period, by
// the date each transaction was entered for: its income and expense by
// category, its balance on each account today, and its budgets.
type CompanyReport struct {
	Company           int64
	From              int64
	To                int64
	Income            int64
	Expense           int64
	PendingIncome     int64
	PendingExpense    int64
	IncomeByCategory  []CategoryTotal
	ExpenseByCategory []CategoryTotal
	Balances          []AccountBalance
	Totals            BalanceTotals
	Budgets           []BudgetLine
}

// ComputeCompanyReport is the report of the company with id company (0 for
// every one) between from and to, over book.
func ComputeCompanyReport(sandbox *api.Sandbox, book Book, company int64, from int64, to int64) CompanyReport {
	report := CompanyReport{Company: company, From: from, To: to}
	totals := map[int64]*CategoryTotal{}
	order := []int64{}
	for _, tx := range book.Transactions {
		if company != 0 && tx.Company != company {
			continue
		}
		if tx.Type != erpledger.TypeIncome && tx.Type != erpledger.TypeExpense {
			continue
		}
		if tx.Date < from || tx.Date > to {
			continue
		}
		key := tx.Category
		if key == 0 {
			key = -1
			if tx.Type == erpledger.TypeExpense {
				key = -2
			}
		}
		total, ok := totals[key]
		if !ok {
			total = &CategoryTotal{Category: tx.Category, Type: tx.Type}
			totals[key] = total
			order = append(order, key)
		}
		amount := tx.Amount
		if tx.Type == erpledger.TypeExpense {
			amount = -amount
		}
		settled := tx.Status == erpledger.StatusSettled
		switch {
		case settled:
			total.Realized += amount
		default:
			total.Pending += amount
		}
		if tx.Type == erpledger.TypeIncome {
			if settled {
				report.Income += amount
			} else {
				report.PendingIncome += amount
			}
		} else if settled {
			report.Expense += amount
		} else {
			report.PendingExpense += amount
		}
	}
	for _, key := range order {
		total := *totals[key]
		if total.Type == erpledger.TypeIncome {
			report.IncomeByCategory = append(report.IncomeByCategory, total)
		} else {
			report.ExpenseByCategory = append(report.ExpenseByCategory, total)
		}
	}
	byAmount := func(list []CategoryTotal) {
		sandbox.Deps.SortDeps.Slice(list, func(i int, j int) bool {
			return list[i].Realized+list[i].Pending > list[j].Realized+list[j].Pending
		})
	}
	byAmount(report.IncomeByCategory)
	byAmount(report.ExpenseByCategory)

	balances := AccountBalances(sandbox, book, book.Today)
	for _, balance := range balances {
		if company == 0 {
			report.Balances = append(report.Balances, balance)
			continue
		}
		if _, ok := balance.ByCompany[company]; ok {
			report.Balances = append(report.Balances, balance)
		}
	}
	report.Totals = TotalsOf(sandbox, balances)
	report.Budgets = ComputeBudgetLines(sandbox, book, erpstore.BudgetQuery{Company: company, From: from, To: to})
	return report
}

// ComputeBudgetLines is every budget query keeps, set against what the
// transactions of its period hold.
func ComputeBudgetLines(sandbox *api.Sandbox, book Book, query erpstore.BudgetQuery) []BudgetLine {
	lines := []BudgetLine{}
	for _, budget := range book.Budgets {
		if query.Company != 0 && budget.Company != query.Company {
			continue
		}
		if query.Category != 0 && budget.Category != query.Category {
			continue
		}
		if (query.From != 0 && budget.PeriodEnd < query.From) || (query.To != 0 && budget.PeriodStart > query.To) {
			continue
		}
		realized, pending := BudgetActual(sandbox, book, budget)
		line := BudgetLine{Budget: budget, Type: book.CategoryById[budget.Category].Type, Realized: realized, Pending: pending}
		line.Committed = realized + pending
		line.Remaining = budget.Amount - line.Committed
		if budget.Amount > 0 {
			line.Percent = line.Committed * 100 / budget.Amount
		}
		lines = append(lines, line)
	}
	return lines
}

// CardSummary is where one credit card stands.
type CardSummary struct {
	Card erp_db.AccountRecord
	// Debt is what the card owes in all, every invoice counted: purchases
	// less refunds and payments, the future installments included.
	Debt int64
	// Available is the credit limit less the debt; 0 when the card has no
	// limit.
	Available int64
	// Current is the invoice still open — the one purchases land on today.
	Current erpledger.Invoice
	// NextDue is the oldest invoice not paid yet, whatever its status.
	NextDue    erpledger.Invoice
	HasNextDue bool
}

// ComputeCardSummaries is where every credit card stands today.
func ComputeCardSummaries(sandbox *api.Sandbox, book Book) []CardSummary {
	summaries := []CardSummary{}
	for _, card := range Cards(sandbox, book) {
		summary := CardSummary{Card: card}
		summary.Debt = -erpledger.Balance(sandbox, book.Transactions, card.Id, 0, 0)
		if card.CreditLimit > 0 {
			summary.Available = card.CreditLimit - summary.Debt
		}
		current := erpledger.InvoiceFor(sandbox, card, book.Today)
		summary.Current = erpledger.FindInvoice(sandbox, card, book.Transactions, current, book.Today)
		for _, invoice := range erpledger.Invoices(sandbox, card, book.Transactions, book.Today) {
			if invoice.Remaining > 0 && invoice.Status != erpledger.InvoiceOpen {
				summary.NextDue, summary.HasNextDue = invoice, true
				break
			}
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

// Dashboard is the overview of every company's finances today.
type Dashboard struct {
	Today    int64
	Balances []AccountBalance
	Totals   BalanceTotals
	// Due30 are the payables and receivables of the next 30 days, the
	// overdue ones included.
	Due30 DueList
	// MonthIncome and MonthExpense are what this month's transactions add
	// up to, settled and pending, by the date they were entered for; the
	// budgets are this month's.
	MonthIncome        int64
	MonthExpense       int64
	MonthBudgetIncome  int64
	MonthBudgetExpense int64
	Cards              []CardSummary
	Projection         Projection
}

// ComputeDashboard is the Dashboard over book, its projection looking months
// ahead.
func ComputeDashboard(sandbox *api.Sandbox, book Book, months int64) Dashboard {
	dashboard := Dashboard{Today: book.Today}
	dashboard.Balances = AccountBalances(sandbox, book, book.Today)
	dashboard.Totals = TotalsOf(sandbox, dashboard.Balances)
	dashboard.Due30 = ComputeDue(sandbox, book, DueQuery{To: erpdate.AddDays(sandbox, book.Today, 30)})
	start, end := erpdate.MonthStart(sandbox, book.Today), erpdate.MonthEnd(sandbox, book.Today)
	month := ComputeCompanyReport(sandbox, book, 0, start, end)
	dashboard.MonthIncome = month.Income + month.PendingIncome
	dashboard.MonthExpense = month.Expense + month.PendingExpense
	for _, line := range month.Budgets {
		if line.Budget.PeriodStart > end || line.Budget.PeriodEnd < start {
			continue
		}
		if line.Type == erpledger.TypeIncome {
			dashboard.MonthBudgetIncome += line.Budget.Amount
		} else {
			dashboard.MonthBudgetExpense += line.Budget.Amount
		}
	}
	dashboard.Cards = ComputeCardSummaries(sandbox, book)
	dashboard.Projection = ComputeProjection(sandbox, book, ProjectionQuery{Months: months})
	return dashboard
}
