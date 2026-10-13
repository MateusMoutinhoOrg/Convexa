package erpview

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpmoney"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// EventView is one event — a movement of money done or expected — as a
// report lists it.
type EventView struct {
	Event       erpreports.Event
	Description string
	Company     string
	Account     string
	Category    string
	ThirdParty  string
	// Href leads to what the event comes from: the transaction, the invoice,
	// the recurrence or the budget.
	Href string
	// DueDate is when it is due, for a payable or a receivable.
	DueDate int64
	// Amount is unsigned, for a payable or a receivable.
	Amount int64
	Kind   string
	// CanSettle offers to settle the transaction it is.
	CanSettle bool
}

// EventViewOf is the view of event, named from book in lang.
func EventViewOf(sandbox *api.Sandbox, book erpreports.Book, lang string, event erpreports.Event) EventView {
	view := EventView{
		Event: event, Description: event.Description, Company: book.CompanyById[event.Company].Name,
		Account: book.AccountById[event.Account].Name, Category: book.CategoryPath[event.Category],
		ThirdParty: book.ThirdPartyById[event.ThirdParty].Name, DueDate: event.Date,
	}
	switch event.Source {
	case erpreports.SourceInvoice:
		card := book.AccountById[event.Card]
		view.Description = card.Name + " · " + erpdate.FormatMonth(sandbox, lang, event.Month)
		view.Href = Href(sandbox, "/admin/erp/card-invoice/"+sandbox.Deps.StringsDeps.FormatInt(event.Card, 10), "month", erpdate.ISOMonth(sandbox, event.Month))
	case erpreports.SourceBudget:
		view.Description = book.CategoryPath[event.Category]
		view.Href = "/admin/erp/list-budgets"
	case erpreports.SourceRecurrence:
		view.Href = "/admin/manage/set-recurrence/" + sandbox.Deps.StringsDeps.FormatInt(event.Recurrence, 10)
	default:
		view.Href = "/admin/manage/set-transaction/" + sandbox.Deps.StringsDeps.FormatInt(event.TxId, 10)
		view.CanSettle = event.Source == erpreports.SourcePending
		if erpledger.IsTransfer(sandbox, event.Type) {
			view.Href = "/admin/manage/set-transfer/" + sandbox.Deps.StringsDeps.FormatInt(event.TxId, 10)
		}
	}
	return view
}

// DueView is one payable or receivable as the page lists it.
type DueView struct {
	EventView
	Item erpreports.DueItem
}

// DuePage is what erp/payables.html is rendered with.
type DuePage struct {
	Kind         string
	Company      string
	Account      string
	ThirdParty   string
	From         string
	To           string
	Overdue      bool
	Kinds        []Option
	Companies    []Option
	Accounts     []Option
	ThirdParties []Option
	List         erpreports.DueList
	Items        []DueView
	Self         string
	TodayISO     string
}

// RenderDuePage answers the payables and receivables the filters keep —
// due up to to (60 days ahead when blank), the overdue ones always.
func RenderDuePage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, kind string, company string, account string, party string, from string, to string, overdue string, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lang := erppage.Language(sandbox, props)
	reader := Reader{}
	query := erpreports.DueQuery{
		Kind: kind, Company: ParseId(sandbox, company), Account: ParseId(sandbox, account), ThirdParty: ParseId(sandbox, party),
		From: Date(sandbox, &reader, "from", from), To: Date(sandbox, &reader, "to", to), OverdueOnly: overdue != "",
	}
	if query.To == 0 {
		query.To = erpdate.AddDays(sandbox, book.Today, 60)
		to = erpdate.ISO(sandbox, query.To)
	}
	list := erpreports.ComputeDue(sandbox, book, query)
	body := DuePage{
		Kind: kind, Company: company, Account: account, ThirdParty: party, From: from, To: to, Overdue: query.OverdueOnly, List: list,
		Kinds:        Choices(sandbox, []string{erpreports.KindPayable, erpreports.KindReceivable}, kind),
		Companies:    CompanyOptions(sandbox, book.Companies, query.Company),
		Accounts:     AccountOptions(sandbox, book.Accounts, query.Account, false),
		ThirdParties: ThirdPartyOptions(sandbox, book.ThirdParties, query.ThirdParty),
		Self:         Href(sandbox, "/admin/erp/payables-receivables", "kind", kind, "company", company, "account", account, "third-party", party, "from", from, "to", to, "overdue", overdue),
		TodayISO:     erpdate.ISO(sandbox, book.Today),
	}
	for _, item := range list.Items {
		view := DueView{EventView: EventViewOf(sandbox, book, lang, item.Event), Item: item}
		view.DueDate, view.Amount, view.Kind = item.DueDate, item.Amount, item.Kind
		body.Items = append(body.Items, view)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/payables.html", "Payables & receivables", "payables", NoticeOf(sandbox, notice), body)
}

// CashFlowPage is what erp/cash_flow.html is rendered with.
type CashFlowPage struct {
	From         string
	To           string
	Group        string
	Basis        string
	Company      string
	Account      string
	Category     string
	ThirdParty   string
	Groups       []Option
	Bases        []Option
	Companies    []Option
	Accounts     []Option
	Categories   []Option
	ThirdParties []Option
	Flow         erpreports.CashFlow
	Labels       []string
	Chart        Chart
}

// periodLabel is how a row of group starting on start is labeled in lang.
func periodLabel(sandbox *api.Sandbox, lang string, group string, start int64, end int64) string {
	switch group {
	case erpreports.GroupMonth:
		return erpdate.FormatMonth(sandbox, lang, erpdate.Month(sandbox, start))
	case erpreports.GroupWeek:
		return erpdate.Format(sandbox, lang, start) + " – " + erpdate.Format(sandbox, lang, end)
	}
	return erpdate.Format(sandbox, lang, start)
}

// RenderCashFlowPage answers the cash flow the filters ask for.
func RenderCashFlowPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, from string, to string, group string, basis string, company string, account string, category string, party string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lang := erppage.Language(sandbox, props)
	reader := Reader{}
	query := erpreports.NormalizeCashFlowQuery(sandbox, erpreports.CashFlowQuery{
		From: Date(sandbox, &reader, "from", from), To: Date(sandbox, &reader, "to", to), Group: group, Basis: basis,
		Company: ParseId(sandbox, company), Account: ParseId(sandbox, account), Category: ParseId(sandbox, category), ThirdParty: ParseId(sandbox, party),
	})
	flow := erpreports.ComputeCashFlow(sandbox, book, query)
	body := CashFlowPage{
		From: erpdate.ISO(sandbox, query.From), To: erpdate.ISO(sandbox, query.To), Group: query.Group, Basis: query.Basis,
		Company: company, Account: account, Category: category, ThirdParty: party, Flow: flow,
		Groups:       Choices(sandbox, []string{erpreports.GroupDay, erpreports.GroupWeek, erpreports.GroupMonth}, query.Group),
		Bases:        Choices(sandbox, []string{erpreports.BasisRealized, erpreports.BasisProjected}, query.Basis),
		Companies:    CompanyOptions(sandbox, book.Companies, query.Company),
		Accounts:     AccountOptions(sandbox, book.Accounts, query.Account, true),
		Categories:   CategoryOptions(sandbox, book.Categories, "", query.Category),
		ThirdParties: ThirdPartyOptions(sandbox, book.ThirdParties, query.ThirdParty),
	}
	groups := [][]int64{}
	for _, row := range flow.Rows {
		body.Labels = append(body.Labels, periodLabel(sandbox, lang, query.Group, row.Start, row.End))
		groups = append(groups, []int64{row.Inflows, -row.Outflows})
	}
	money := func(cents int64) string { return CompactMoney(sandbox, cents) }
	body.Chart = BarChart(sandbox, 960, 260, groups, []string{"bar-in", "bar-out"}, money,
		func(i int) string {
			row := flow.Rows[i]
			if query.Group == erpreports.GroupMonth {
				return erpdate.FormatMonth(sandbox, lang, erpdate.Month(sandbox, row.Start))
			}
			_, month, day := erpdate.Split(sandbox, row.Start)
			if lang == erpi18n.Portuguese {
				return sandbox.Deps.StdDeps.Sprintf("%02d/%02d", day, month)
			}
			return sandbox.Deps.StdDeps.Sprintf("%s %d", erpdate.MonthName(sandbox, lang, month), day)
		},
		func(g int, s int) string { return body.Labels[g] + ": " + erpmoney.Format(sandbox, lang, groups[g][s]) })
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/cash_flow.html", "Cash flow", "cash-flow", erppage.Notice{}, body)
}

// ProjectionPage is what erp/projection.html is rendered with.
type ProjectionPage struct {
	Months     string
	Company    string
	Account    string
	MonthsList []Option
	Companies  []Option
	Accounts   []Option
	Projection erpreports.Projection
	Chart      Chart
	Events     []EventView
	// MonthRows are the months of the projection, labeled.
	MonthRows []ProjectionMonthView
}

// ProjectionMonthView is one month of a projection, labeled.
type ProjectionMonthView struct {
	Month erpreports.ProjectionMonth
	Label string
}

// ProjectionChart is the chart of the daily balances of projection, with
// and without the budgets, in lang.
func ProjectionChart(sandbox *api.Sandbox, lang string, projection erpreports.Projection, width int64, height int64) Chart {
	plain, budgeted := []int64{}, []int64{}
	for _, point := range projection.Points {
		plain = append(plain, point.Balance)
		budgeted = append(budgeted, point.WithBudgets)
	}
	series := [][]int64{plain}
	classes := []string{"line-a"}
	if projection.HasBudgets {
		series = append(series, budgeted)
		classes = append(classes, "line-b")
	}
	return LineChart(sandbox, width, height, series, classes, func(cents int64) string { return CompactMoney(sandbox, cents) },
		func(i int) string {
			date := projection.Points[i].Date
			return erpdate.FormatMonth(sandbox, lang, erpdate.Month(sandbox, date))
		})
}

// RenderProjectionPage answers the projected balance the filters ask for.
func RenderProjectionPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, months string, company string, account string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lang := erppage.Language(sandbox, props)
	query := erpreports.ProjectionQuery{Months: ParseId(sandbox, months), Company: ParseId(sandbox, company), Account: ParseId(sandbox, account)}
	projection := erpreports.ComputeProjection(sandbox, book, query)
	monthChoices := []Option{}
	for _, count := range []int64{3, 6, 12, 18, 24, 36} {
		value := sandbox.Deps.StringsDeps.FormatInt(count, 10)
		monthChoices = append(monthChoices, Option{Value: value, Label: value, Selected: count == projection.Query.Months})
	}
	body := ProjectionPage{
		Months: sandbox.Deps.StringsDeps.FormatInt(projection.Query.Months, 10), Company: company, Account: account,
		MonthsList: monthChoices, Projection: projection,
		Companies: CompanyOptions(sandbox, book.Companies, query.Company),
		Accounts:  AccountOptions(sandbox, book.Accounts, query.Account, false),
		Chart:     ProjectionChart(sandbox, lang, projection, 960, 300),
	}
	for _, month := range projection.Months {
		body.MonthRows = append(body.MonthRows, ProjectionMonthView{Month: month, Label: erpdate.FormatMonth(sandbox, lang, month.Month)})
	}
	for _, event := range projection.Events {
		body.Events = append(body.Events, EventViewOf(sandbox, book, lang, event))
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/projection.html", "Projected balance", "projection", erppage.Notice{}, body)
}

// CategoryTotalView is one category of a company report, named.
type CategoryTotalView struct {
	Total erpreports.CategoryTotal
	Name  string
	// Share is the category's part of its type's total, in percent.
	Share int64
}

// CompanyReportPage is what erp/company_report.html is rendered with.
type CompanyReportPage struct {
	Company   string
	From      string
	To        string
	Companies []Option
	Report    erpreports.CompanyReport
	Name      string
	Income    []CategoryTotalView
	Expense   []CategoryTotalView
	Accounts  []CompanyAccountView
	Budgets   []BudgetView
	// Cash and Debt are the company's — or every company's — today.
	Cash int64
	Debt int64
	// Result is the income less the expense, the pending ones included.
	Result int64
}

// CompanyAccountView is the share of a company in one account.
type CompanyAccountView struct {
	Account string
	IsCard  bool
	Amount  int64
	Total   int64
}

// totalsViews are the views of totals, named from book.
func totalsViews(sandbox *api.Sandbox, book erpreports.Book, totals []erpreports.CategoryTotal) []CategoryTotalView {
	sum := int64(0)
	for _, total := range totals {
		sum += total.Realized + total.Pending
	}
	views := []CategoryTotalView{}
	for _, total := range totals {
		name := book.CategoryPath[total.Category]
		if total.Category == 0 {
			name = "No category"
		}
		view := CategoryTotalView{Total: total, Name: name}
		if sum > 0 {
			view.Share = (total.Realized + total.Pending) * 100 / sum
		}
		views = append(views, view)
	}
	return views
}

// RenderCompanyReportPage answers the report of company (every one when
// blank) between from and to (this month when blank).
func RenderCompanyReportPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, company string, from string, to string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	reader := Reader{}
	start, end := Date(sandbox, &reader, "from", from), Date(sandbox, &reader, "to", to)
	if start == 0 {
		start = erpdate.MonthStart(sandbox, book.Today)
	}
	if end == 0 || end < start {
		end = erpdate.MonthEnd(sandbox, start)
	}
	id := ParseId(sandbox, company)
	report := erpreports.ComputeCompanyReport(sandbox, book, id, start, end)
	body := CompanyReportPage{
		Company: company, From: erpdate.ISO(sandbox, start), To: erpdate.ISO(sandbox, end),
		Companies: CompanyOptions(sandbox, book.Companies, id), Report: report, Name: book.CompanyById[id].Name,
		Income: totalsViews(sandbox, book, report.IncomeByCategory), Expense: totalsViews(sandbox, book, report.ExpenseByCategory),
		Budgets: BudgetViews(sandbox, book, report.Budgets),
		Result:  report.Income + report.PendingIncome - report.Expense - report.PendingExpense,
	}
	for _, balance := range report.Balances {
		view := CompanyAccountView{Account: balance.Account.Name, IsCard: erpledger.IsCard(sandbox, balance.Account), Total: balance.Total, Amount: balance.Total}
		if id != 0 {
			view.Amount = balance.ByCompany[id]
		}
		if view.IsCard {
			body.Debt -= view.Amount
		} else {
			body.Cash += view.Amount
		}
		body.Accounts = append(body.Accounts, view)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/company_report.html", "By company", "company-report", erppage.Notice{}, body)
}

// MatrixRow is one account of the balance matrix: its balance whole and in
// each company's column.
type MatrixRow struct {
	Account erp_db.AccountRecord
	IsCard  bool
	Total   int64
	Cells   []MatrixCell
}

// MatrixCell is one company's share of one account, Present false when the
// company has no transaction on it.
type MatrixCell struct {
	Amount  int64
	Present bool
}

// AccountReportPage is what erp/account_report.html is rendered with.
type AccountReportPage struct {
	Date      string
	Account   string
	Company   string
	From      string
	To        string
	Accounts  []Option
	Companies []Option
	// Columns are the companies of the matrix.
	Columns []string
	Rows    []MatrixRow
	// ColumnCash totals each company's cash; ColumnDebt its card debt.
	ColumnCash []int64
	ColumnDebt []int64
	Cash       int64
	Debt       int64
	// HasStatement is true when an account is chosen.
	HasStatement bool
	Statement    erpreports.Statement
	Lines        []StatementView
}

// StatementView is one line of a statement, named.
type StatementView struct {
	Line    erpreports.StatementLine
	Row     TxRow
	Balance int64
}

// RenderAccountReportPage answers the balance of every account by company
// at date (today when blank) and, with an account, its statement between
// from and to (this month when blank).
func RenderAccountReportPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, date string, account string, company string, from string, to string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	reader := Reader{}
	at := Date(sandbox, &reader, "date", date)
	if at == 0 {
		at = book.Today
	}
	accountId, companyId := ParseId(sandbox, account), ParseId(sandbox, company)
	body := AccountReportPage{
		Date: erpdate.ISO(sandbox, at), Account: account, Company: company,
		Accounts:  AccountOptions(sandbox, book.Accounts, accountId, true),
		Companies: CompanyOptions(sandbox, book.Companies, companyId),
	}
	balances := erpreports.AccountBalances(sandbox, book, at)
	columns := []int64{}
	for _, item := range book.Companies {
		for _, balance := range balances {
			if _, ok := balance.ByCompany[item.Id]; ok {
				columns = append(columns, item.Id)
				body.Columns = append(body.Columns, item.Name)
				break
			}
		}
	}
	body.ColumnCash = make([]int64, len(columns))
	body.ColumnDebt = make([]int64, len(columns))
	for _, balance := range balances {
		row := MatrixRow{Account: balance.Account, IsCard: erpledger.IsCard(sandbox, balance.Account), Total: balance.Total}
		for index, column := range columns {
			amount, ok := balance.ByCompany[column]
			row.Cells = append(row.Cells, MatrixCell{Amount: amount, Present: ok})
			if row.IsCard {
				body.ColumnDebt[index] -= amount
			} else {
				body.ColumnCash[index] += amount
			}
		}
		if row.IsCard {
			body.Debt -= balance.Total
		} else {
			body.Cash += balance.Total
		}
		body.Rows = append(body.Rows, row)
	}
	if _, ok := book.AccountById[accountId]; ok {
		start, end := Date(sandbox, &reader, "from", from), Date(sandbox, &reader, "to", to)
		if start == 0 {
			start = erpdate.MonthStart(sandbox, book.Today)
		}
		if end == 0 || end < start {
			end = erpdate.MonthEnd(sandbox, start)
		}
		body.HasStatement = true
		body.From, body.To = erpdate.ISO(sandbox, start), erpdate.ISO(sandbox, end)
		body.Statement = erpreports.AccountStatement(sandbox, book, accountId, companyId, start, end)
		txs := []erp_db.TransactionRecord{}
		for _, line := range body.Statement.Lines {
			txs = append(txs, line.Tx)
		}
		rows := Rows(sandbox, book, txs)
		for index, line := range body.Statement.Lines {
			body.Lines = append(body.Lines, StatementView{Line: line, Row: rows[index], Balance: line.Balance})
		}
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/account_report.html", "By account", "account-report", erppage.Notice{}, body)
}

// DashboardPage is what erp/dashboard.html is rendered with.
type DashboardPage struct {
	Dashboard  erpreports.Dashboard
	Accounts   []AccountRow
	Companies  []CompanyRow
	Cards      []CardView
	Upcoming   []DueView
	Chart      Chart
	MonthLabel string
	// MonthNet is this month's income less its expense.
	MonthNet int64
	// Net is the cash less the card debt.
	Net int64
	// ExpenseMeter is this month's expense against its budget, 0 to 100.
	ExpenseMeter int64
	ExpenseOver  bool
	Empty        bool
}

// RenderDashboardPage answers the dashboard.
func RenderDashboardPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lang := erppage.Language(sandbox, props)
	dashboard := erpreports.ComputeDashboard(sandbox, book, 12)
	body := DashboardPage{
		Dashboard: dashboard, MonthLabel: erpdate.FormatMonth(sandbox, lang, erpdate.Month(sandbox, book.Today)),
		MonthNet: dashboard.MonthIncome - dashboard.MonthExpense, Net: dashboard.Totals.Cash - dashboard.Totals.CardDebt,
		Chart: ProjectionChart(sandbox, lang, dashboard.Projection, 560, 260),
		Empty: len(book.Companies) == 0 || len(book.Accounts) == 0,
	}
	if dashboard.MonthBudgetExpense > 0 {
		body.ExpenseMeter = dashboard.MonthExpense * 100 / dashboard.MonthBudgetExpense
		body.ExpenseOver = body.ExpenseMeter > 100
		if body.ExpenseMeter > 100 {
			body.ExpenseMeter = 100
		}
	}
	for _, balance := range dashboard.Balances {
		if erpledger.IsCard(sandbox, balance.Account) || (balance.Total == 0 && balance.Pending == 0) {
			continue
		}
		body.Accounts = append(body.Accounts, AccountRow{Account: balance.Account, Balance: balance, Shares: SharesOf(sandbox, book, balance)})
	}
	for _, company := range book.Companies {
		if company.Status != erpstore.StatusActive {
			continue
		}
		body.Companies = append(body.Companies, CompanyRow{Company: company, Cash: dashboard.Totals.CashByCompany[company.Id], Debt: dashboard.Totals.DebtByCompany[company.Id]})
	}
	for _, summary := range dashboard.Cards {
		body.Cards = append(body.Cards, CardView{Summary: summary, PaymentAccount: book.AccountById[summary.Card.PaymentAccount].Name})
	}
	for index, item := range dashboard.Due30.Items {
		if index >= 12 {
			break
		}
		view := DueView{EventView: EventViewOf(sandbox, book, lang, item.Event), Item: item}
		view.DueDate, view.Amount, view.Kind = item.DueDate, item.Amount, item.Kind
		body.Upcoming = append(body.Upcoming, view)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/dashboard.html", "Dashboard", "dashboard", erppage.Notice{}, body)
}
