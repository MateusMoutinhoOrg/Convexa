package erpapi

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
)

// The documents of the reports: the JSON twins of the report pages.

// BalancesJSON is the balance of every account of balances — whole, pending
// and by company — and the totals of the cash and of the card debt.
func BalancesJSON(sandbox *api.Sandbox, balances []erpreports.AccountBalance) *serializabledeps.SerializableObject {
	document := Object(sandbox)
	accounts := Array(sandbox)
	for _, balance := range balances {
		accounts.AddItemToArray(AccountJSON(sandbox, balance.Account, balance.Total, balance.Pending, balance.ByCompany))
	}
	totals := erpreports.TotalsOf(sandbox, balances)
	document.AddItemToObject("accounts", accounts)
	document.AddItemToObject("cash", totals.Cash)
	document.AddItemToObject("card-debt", totals.CardDebt)
	document.AddItemToObject("cash-by-company", byCompanyJSON(sandbox, totals.CashByCompany))
	document.AddItemToObject("card-debt-by-company", byCompanyJSON(sandbox, totals.DebtByCompany))
	return document
}

// byCompanyJSON is [{"company", "amount"}] of amounts, by company id.
func byCompanyJSON(sandbox *api.Sandbox, amounts map[int64]int64) *serializabledeps.SerializableObject {
	ids := []int64{}
	for id := range amounts {
		ids = append(ids, id)
	}
	sandbox.Deps.SortDeps.Slice(ids, func(i int, j int) bool { return ids[i] < ids[j] })
	array := Array(sandbox)
	for _, id := range ids {
		item := Object(sandbox)
		item.AddItemToObject("company", id)
		item.AddItemToObject("amount", amounts[id])
		array.AddItemToArray(item)
	}
	return array
}

// EventJSON is one event: money done or expected to move.
func EventJSON(sandbox *api.Sandbox, event erpreports.Event) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("date", erpdate.ISO(sandbox, event.Date))
	object.AddItemToObject("amount", event.Amount)
	object.AddItemToObject("source", event.Source)
	object.AddItemToObject("type", event.Type)
	object.AddItemToObject("overdue", event.Overdue)
	object.AddItemToObject("description", event.Description)
	object.AddItemToObject("company", event.Company)
	object.AddItemToObject("account", event.Account)
	object.AddItemToObject("category", event.Category)
	object.AddItemToObject("third-party", event.ThirdParty)
	object.AddItemToObject("transaction", event.TxId)
	object.AddItemToObject("recurrence", event.Recurrence)
	object.AddItemToObject("card", event.Card)
	object.AddItemToObject("card-invoice", erpdate.ISOMonth(sandbox, event.Month))
	object.AddItemToObject("budget", event.Budget)
	return object
}

// DueJSON is the payables and receivables of list and their totals.
func DueJSON(sandbox *api.Sandbox, list erpreports.DueList) *serializabledeps.SerializableObject {
	items := Array(sandbox)
	for _, item := range list.Items {
		object := EventJSON(sandbox, item.Event)
		object.AddItemToObject("kind", item.Kind)
		object.AddItemToObject("due-date", erpdate.ISO(sandbox, item.DueDate))
		object.ReplaceItemInObject("amount", item.Amount)
		items.AddItemToArray(object)
	}
	document := Object(sandbox)
	document.AddItemToObject("items", items)
	document.AddItemToObject("payable", list.Payable)
	document.AddItemToObject("receivable", list.Receivable)
	document.AddItemToObject("overdue-payable", list.OverduePayable)
	document.AddItemToObject("overdue-receivable", list.OverdueReceivable)
	return document
}

// CashFlowJSON is a cash flow.
func CashFlowJSON(sandbox *api.Sandbox, flow erpreports.CashFlow) *serializabledeps.SerializableObject {
	rows := Array(sandbox)
	for _, row := range flow.Rows {
		object := Object(sandbox)
		object.AddItemToObject("start", erpdate.ISO(sandbox, row.Start))
		object.AddItemToObject("end", erpdate.ISO(sandbox, row.End))
		object.AddItemToObject("inflows", row.Inflows)
		object.AddItemToObject("outflows", row.Outflows)
		object.AddItemToObject("net", row.Net)
		if flow.HasBalance {
			object.AddItemToObject("balance", row.Balance)
		}
		rows.AddItemToArray(object)
	}
	document := Object(sandbox)
	document.AddItemToObject("from", erpdate.ISO(sandbox, flow.Query.From))
	document.AddItemToObject("to", erpdate.ISO(sandbox, flow.Query.To))
	document.AddItemToObject("group", flow.Query.Group)
	document.AddItemToObject("basis", flow.Query.Basis)
	document.AddItemToObject("has-balance", flow.HasBalance)
	document.AddItemToObject("opening", flow.Opening)
	document.AddItemToObject("inflows", flow.Inflows)
	document.AddItemToObject("outflows", flow.Outflows)
	document.AddItemToObject("net", flow.Net)
	document.AddItemToObject("closing", flow.Closing)
	document.AddItemToObject("rows", rows)
	return document
}

// ProjectionJSON is a projection, its events listed when events is true.
func ProjectionJSON(sandbox *api.Sandbox, projection erpreports.Projection, events bool) *serializabledeps.SerializableObject {
	months := Array(sandbox)
	for _, month := range projection.Months {
		object := Object(sandbox)
		object.AddItemToObject("month", erpdate.ISOMonth(sandbox, month.Month))
		object.AddItemToObject("opening", month.Opening)
		object.AddItemToObject("income", month.Income)
		object.AddItemToObject("expense", month.Expense)
		object.AddItemToObject("closing", month.Closing)
		if projection.HasBudgets {
			object.AddItemToObject("budgets", month.Budgets)
			object.AddItemToObject("closing-with-budgets", month.ClosingWithBudgets)
		}
		months.AddItemToArray(object)
	}
	document := Object(sandbox)
	document.AddItemToObject("start", erpdate.ISO(sandbox, projection.Start))
	document.AddItemToObject("end", erpdate.ISO(sandbox, projection.End))
	document.AddItemToObject("opening", projection.Opening)
	document.AddItemToObject("income", projection.Income)
	document.AddItemToObject("expense", projection.Expense)
	document.AddItemToObject("final", projection.Final)
	document.AddItemToObject("lowest", projection.Lowest)
	document.AddItemToObject("lowest-date", erpdate.ISO(sandbox, projection.LowestDate))
	document.AddItemToObject("has-budgets", projection.HasBudgets)
	if projection.HasBudgets {
		document.AddItemToObject("budgets", projection.BudgetTotal)
		document.AddItemToObject("final-with-budgets", projection.FinalWithBudgets)
		document.AddItemToObject("lowest-with-budgets", projection.LowestWithBudgets)
		document.AddItemToObject("lowest-with-budgets-date", erpdate.ISO(sandbox, projection.LowestWithBudgetsDate))
	}
	document.AddItemToObject("months", months)
	if events {
		list := Array(sandbox)
		for _, event := range projection.Events {
			list.AddItemToArray(EventJSON(sandbox, event))
		}
		document.AddItemToObject("events", list)
	}
	return document
}

// BudgetLinesJSON is the budgets of lines set against what happened.
func BudgetLinesJSON(sandbox *api.Sandbox, lines []erpreports.BudgetLine) *serializabledeps.SerializableObject {
	array := Array(sandbox)
	for _, line := range lines {
		object := BudgetJSON(sandbox, line.Budget)
		object.AddItemToObject("type", line.Type)
		object.AddItemToObject("realized", line.Realized)
		object.AddItemToObject("pending", line.Pending)
		object.AddItemToObject("committed", line.Committed)
		object.AddItemToObject("remaining", line.Remaining)
		object.AddItemToObject("percent", line.Percent)
		array.AddItemToArray(object)
	}
	return array
}

// categoryTotalsJSON is totals, each with its category's full name in paths.
func categoryTotalsJSON(sandbox *api.Sandbox, totals []erpreports.CategoryTotal, paths map[int64]string) *serializabledeps.SerializableObject {
	array := Array(sandbox)
	for _, total := range totals {
		object := Object(sandbox)
		object.AddItemToObject("category", total.Category)
		object.AddItemToObject("path", paths[total.Category])
		object.AddItemToObject("realized", total.Realized)
		object.AddItemToObject("pending", total.Pending)
		array.AddItemToArray(object)
	}
	return array
}

// CompanyReportJSON is a company report, the categories named from book.
func CompanyReportJSON(sandbox *api.Sandbox, book erpreports.Book, report erpreports.CompanyReport) *serializabledeps.SerializableObject {
	document := Object(sandbox)
	document.AddItemToObject("company", report.Company)
	document.AddItemToObject("from", erpdate.ISO(sandbox, report.From))
	document.AddItemToObject("to", erpdate.ISO(sandbox, report.To))
	document.AddItemToObject("income", report.Income)
	document.AddItemToObject("expense", report.Expense)
	document.AddItemToObject("pending-income", report.PendingIncome)
	document.AddItemToObject("pending-expense", report.PendingExpense)
	document.AddItemToObject("income-by-category", categoryTotalsJSON(sandbox, report.IncomeByCategory, book.CategoryPath))
	document.AddItemToObject("expense-by-category", categoryTotalsJSON(sandbox, report.ExpenseByCategory, book.CategoryPath))
	balances := Array(sandbox)
	for _, balance := range report.Balances {
		object := Object(sandbox)
		object.AddItemToObject("account", balance.Account.Id)
		object.AddItemToObject("kind", balance.Account.Kind)
		share := balance.Total
		if report.Company != 0 {
			share = balance.ByCompany[report.Company]
		}
		object.AddItemToObject("balance", share)
		object.AddItemToObject("account-balance", balance.Total)
		balances.AddItemToArray(object)
	}
	document.AddItemToObject("balances", balances)
	document.AddItemToObject("budgets", BudgetLinesJSON(sandbox, report.Budgets))
	return document
}

// StatementJSON is a statement.
func StatementJSON(sandbox *api.Sandbox, statement erpreports.Statement, today int64) *serializabledeps.SerializableObject {
	lines := Array(sandbox)
	for _, line := range statement.Lines {
		object := TransactionJSON(sandbox, line.Tx, today)
		object.AddItemToObject("balance", line.Balance)
		lines.AddItemToArray(object)
	}
	document := Object(sandbox)
	document.AddItemToObject("account", statement.Account.Id)
	document.AddItemToObject("company", statement.Company)
	document.AddItemToObject("from", erpdate.ISO(sandbox, statement.From))
	document.AddItemToObject("to", erpdate.ISO(sandbox, statement.To))
	document.AddItemToObject("opening", statement.Opening)
	document.AddItemToObject("inflows", statement.Inflows)
	document.AddItemToObject("outflows", statement.Outflows)
	document.AddItemToObject("closing", statement.Closing)
	document.AddItemToObject("lines", lines)
	return document
}

// CardSummaryJSON is where a credit card stands.
func CardSummaryJSON(sandbox *api.Sandbox, summary erpreports.CardSummary) *serializabledeps.SerializableObject {
	object := AccountJSON(sandbox, summary.Card, 0, 0, nil)
	object.AddItemToObject("debt", summary.Debt)
	object.AddItemToObject("available", summary.Available)
	object.AddItemToObject("open-invoice", InvoiceJSON(sandbox, summary.Card.Id, summary.Current))
	if summary.HasNextDue {
		object.AddItemToObject("next-due-invoice", InvoiceJSON(sandbox, summary.Card.Id, summary.NextDue))
	}
	return object
}

// DashboardJSON is the dashboard.
func DashboardJSON(sandbox *api.Sandbox, dashboard erpreports.Dashboard) *serializabledeps.SerializableObject {
	document := BalancesJSON(sandbox, dashboard.Balances)
	document.AddItemToObject("today", erpdate.ISO(sandbox, dashboard.Today))
	document.AddItemToObject("net", dashboard.Totals.Cash-dashboard.Totals.CardDebt)
	document.AddItemToObject("due-30-days", DueJSON(sandbox, dashboard.Due30))
	month := Object(sandbox)
	month.AddItemToObject("income", dashboard.MonthIncome)
	month.AddItemToObject("expense", dashboard.MonthExpense)
	month.AddItemToObject("income-budget", dashboard.MonthBudgetIncome)
	month.AddItemToObject("expense-budget", dashboard.MonthBudgetExpense)
	document.AddItemToObject("this-month", month)
	cards := Array(sandbox)
	for _, summary := range dashboard.Cards {
		cards.AddItemToArray(CardSummaryJSON(sandbox, summary))
	}
	document.AddItemToObject("cards", cards)
	document.AddItemToObject("projection", ProjectionJSON(sandbox, dashboard.Projection, false))
	return document
}
