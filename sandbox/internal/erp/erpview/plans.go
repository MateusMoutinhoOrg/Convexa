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

// The pages of what is planned: the recurrences and the budgets.

// RecurrenceRow is one recurrence of the list.
type RecurrenceRow struct {
	Recurrence erp_db.RecurrenceRecord
	Company    string
	Account    string
	Category   string
	ThirdParty string
	Next       int64
	// Amount is signed, the way the recurrence moves money.
	Amount int64
}

// RecurrencesPage is what erp/recurrences.html is rendered with.
type RecurrencesPage struct {
	Search    string
	Company   string
	Account   string
	Type      string
	Status    string
	Companies []Option
	Accounts  []Option
	Types     []Option
	Statuses  []Option
	Rows      []RecurrenceRow
	// Monthly is what the active ones move in a month, as if every one were
	// monthly: a weekly one counts 52/12 times, a yearly one 1/12.
	MonthlyIncome  int64
	MonthlyExpense int64
	HorizonDays    int64
}

// RenderRecurrencesPage answers the list of the recurrences query keeps.
func RenderRecurrencesPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, search string, company string, account string, kind string, status string, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	recurrences, err := erpstore.Recurrences(sandbox, erpstore.RecurrenceQuery{Search: search, Company: ParseId(sandbox, company), Account: ParseId(sandbox, account), Type: kind, Status: status})
	if err != nil {
		return err
	}
	body := RecurrencesPage{
		Search: search, Company: company, Account: account, Type: kind, Status: status,
		Companies:   CompanyOptions(sandbox, book.Companies, ParseId(sandbox, company)),
		Accounts:    AccountOptions(sandbox, book.Accounts, ParseId(sandbox, account), true),
		Types:       Choices(sandbox, []string{erpledger.TypeIncome, erpledger.TypeExpense}, kind),
		Statuses:    Choices(sandbox, []string{erpstore.RecurrenceActive, erpstore.RecurrencePaused}, status),
		HorizonDays: sandbox.Config.RecurrenceHorizonDays,
	}
	for _, recurrence := range recurrences {
		row := RecurrenceRow{
			Recurrence: recurrence, Company: book.CompanyById[recurrence.Company].Name, Account: book.AccountById[recurrence.Account].Name,
			Category: book.CategoryPath[recurrence.Category], ThirdParty: book.ThirdPartyById[recurrence.ThirdParty].Name,
			Next: erpstore.NextOccurrence(sandbox, recurrence, erpdate.AddDays(sandbox, book.Today, -1)), Amount: recurrence.Amount,
		}
		if recurrence.Type == erpledger.TypeExpense {
			row.Amount = -row.Amount
		}
		if recurrence.Status == erpstore.RecurrenceActive && row.Next != 0 {
			monthly := recurrence.Amount
			switch recurrence.Frequency {
			case erpstore.FrequencyWeekly:
				monthly = recurrence.Amount * 52 / 12
			case erpstore.FrequencyYearly:
				monthly = recurrence.Amount / 12
			}
			monthly /= recurrence.Interval
			if recurrence.Type == erpledger.TypeIncome {
				body.MonthlyIncome += monthly
			} else {
				body.MonthlyExpense += monthly
			}
		}
		body.Rows = append(body.Rows, row)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/recurrences.html", "Recurrences", "recurrences", NoticeOf(sandbox, notice), body)
}

// RecurrenceForm is what the recurrence form posts, as it posts it.
type RecurrenceForm struct {
	Company     string
	Account     string
	Category    string
	ThirdParty  string
	Type        string
	Amount      string
	Description string
	Notes       string
	Frequency   string
	Interval    string
	StartDate   string
	EndDate     string
	Status      string
}

// RecurrenceFieldsOf reads the RecurrenceFields form posted.
func RecurrenceFieldsOf(sandbox *api.Sandbox, form RecurrenceForm) (erpstore.RecurrenceFields, erpstore.Problem) {
	reader := Reader{}
	fields := erpstore.RecurrenceFields{
		Company: ParseId(sandbox, form.Company), Account: ParseId(sandbox, form.Account), Category: ParseId(sandbox, form.Category),
		ThirdParty: ParseId(sandbox, form.ThirdParty), Type: form.Type, Amount: Money(sandbox, &reader, "amount", form.Amount),
		Description: form.Description, Notes: form.Notes, Frequency: form.Frequency, Interval: Int(sandbox, &reader, "interval", form.Interval),
		StartDate: Date(sandbox, &reader, "start-date", form.StartDate), EndDate: Date(sandbox, &reader, "end-date", form.EndDate), Status: form.Status,
	}
	if fields.Amount < 0 {
		fields.Amount = -fields.Amount
	}
	return fields, ProblemOf(sandbox, reader)
}

// RecurrenceFormOf is the RecurrenceForm that fills the form with fields.
func RecurrenceFormOf(sandbox *api.Sandbox, lang string, fields erpstore.RecurrenceFields) RecurrenceForm {
	form := RecurrenceForm{
		Company: id(sandbox, fields.Company), Account: id(sandbox, fields.Account), Category: id(sandbox, fields.Category),
		ThirdParty: id(sandbox, fields.ThirdParty), Type: fields.Type, Amount: moneyInput(sandbox, lang, fields.Amount),
		Description: fields.Description, Notes: fields.Notes, Frequency: fields.Frequency,
		StartDate: erpdate.ISO(sandbox, fields.StartDate), EndDate: erpdate.ISO(sandbox, fields.EndDate), Status: fields.Status,
	}
	if fields.Interval > 0 {
		form.Interval = sandbox.Deps.StringsDeps.FormatInt(fields.Interval, 10)
	}
	return form
}

// RecurrenceFormPage is what erp/recurrence_form.html is rendered with.
type RecurrenceFormPage struct {
	Form         Form
	Values       RecurrenceForm
	Types        []Option
	Frequencies  []Option
	Statuses     []Option
	Companies    []Option
	Accounts     []Option
	Categories   []Option
	ThirdParties []Option
	HorizonDays  int64
}

// RenderRecurrenceForm answers the form adding a recurrence (id 0) or
// editing the one with id id, filled with values, problem above it.
func RenderRecurrenceForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, values RecurrenceForm, problem erpstore.Problem) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	if values.Type == "" {
		values.Type = erpledger.TypeExpense
	}
	if values.Frequency == "" {
		values.Frequency = erpstore.FrequencyMonthly
	}
	if values.Interval == "" {
		values.Interval = "1"
	}
	if values.StartDate == "" {
		values.StartDate = erpdate.ISO(sandbox, book.Today)
	}
	if values.Status == "" {
		values.Status = erpstore.RecurrenceActive
	}
	form := newForm(sandbox, id, "recurrence", "recurrence", "/admin/erp/list-recurrences", problem)
	body := RecurrenceFormPage{
		Form: form, Values: values, HorizonDays: sandbox.Config.RecurrenceHorizonDays,
		Types:        Choices(sandbox, []string{erpledger.TypeExpense, erpledger.TypeIncome}, values.Type),
		Frequencies:  Choices(sandbox, []string{erpstore.FrequencyWeekly, erpstore.FrequencyMonthly, erpstore.FrequencyYearly}, values.Frequency),
		Statuses:     Choices(sandbox, []string{erpstore.RecurrenceActive, erpstore.RecurrencePaused}, values.Status),
		Companies:    CompanyOptions(sandbox, book.Companies, ParseId(sandbox, values.Company)),
		Accounts:     AccountOptions(sandbox, book.Accounts, ParseId(sandbox, values.Account), true),
		Categories:   CategoryOptions(sandbox, book.Categories, "", ParseId(sandbox, values.Category)),
		ThirdParties: ThirdPartyOptions(sandbox, book.ThirdParties, ParseId(sandbox, values.ThirdParty)),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/recurrence_form.html", "recurrences", form, body)
}

// BudgetsPage is what erp/budgets.html is rendered with.
type BudgetsPage struct {
	Company    string
	Category   string
	From       string
	To         string
	Companies  []Option
	Categories []Option
	Lines      []BudgetView
	// The totals of the lines, by type.
	ExpenseBudget    int64
	ExpenseCommitted int64
	IncomeBudget     int64
	IncomeCommitted  int64
}

// BudgetView is one budget set against what happened in its period.
type BudgetView struct {
	Line     erpreports.BudgetLine
	Company  string
	Category string
	// Meter is the committed share of the budget, 0 to 100, for the bar.
	Meter int64
	Over  bool
}

// BudgetViews are the views of lines, named from book.
func BudgetViews(sandbox *api.Sandbox, book erpreports.Book, lines []erpreports.BudgetLine) []BudgetView {
	views := []BudgetView{}
	for _, line := range lines {
		view := BudgetView{Line: line, Company: book.CompanyById[line.Budget.Company].Name, Category: book.CategoryPath[line.Budget.Category], Meter: line.Percent}
		if view.Meter > 100 {
			view.Meter = 100
		}
		view.Over = line.Type == erpledger.TypeExpense && line.Committed > line.Budget.Amount
		views = append(views, view)
	}
	return views
}

// RenderBudgetsPage answers the budgets of company and category whose period
// overlaps from to to — months spelled 2026-10, this year when blank —
// set against what happened, on the template at template.
func RenderBudgetsPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, template string, active string, title string, company string, category string, from string, to string, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	fromMonth, ok := erpdate.ParseMonth(sandbox, from)
	if !ok {
		fromMonth = erpdate.Month(sandbox, book.Today)/100*100 + 1
	}
	toMonth, ok := erpdate.ParseMonth(sandbox, to)
	if !ok || toMonth < fromMonth {
		toMonth = fromMonth/100*100 + 12
	}
	query := erpstore.BudgetQuery{Company: ParseId(sandbox, company), Category: ParseId(sandbox, category), From: fromMonth*100 + 1, To: erpdate.MonthEnd(sandbox, toMonth*100+1)}
	body := BudgetsPage{
		Company: company, Category: category, From: erpdate.ISOMonth(sandbox, fromMonth), To: erpdate.ISOMonth(sandbox, toMonth),
		Companies:  CompanyOptions(sandbox, book.Companies, query.Company),
		Categories: CategoryOptions(sandbox, book.Categories, "", query.Category),
		Lines:      BudgetViews(sandbox, book, erpreports.ComputeBudgetLines(sandbox, book, query)),
	}
	for _, view := range body.Lines {
		if view.Line.Type == erpledger.TypeIncome {
			body.IncomeBudget += view.Line.Budget.Amount
			body.IncomeCommitted += view.Line.Committed
		} else {
			body.ExpenseBudget += view.Line.Budget.Amount
			body.ExpenseCommitted += view.Line.Committed
		}
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, template, title, active, NoticeOf(sandbox, notice), body)
}

// BudgetForm is what the budget form posts, as it posts it.
type BudgetForm struct {
	Company     string
	Category    string
	PeriodStart string
	PeriodEnd   string
	Amount      string
	Notes       string
	Months      string
}

// BudgetFieldsOf reads the BudgetFields form posted.
func BudgetFieldsOf(sandbox *api.Sandbox, form BudgetForm) (erpstore.BudgetFields, erpstore.Problem) {
	reader := Reader{}
	fields := erpstore.BudgetFields{
		Company: ParseId(sandbox, form.Company), Category: ParseId(sandbox, form.Category),
		PeriodStart: Date(sandbox, &reader, "period-start", form.PeriodStart), PeriodEnd: Date(sandbox, &reader, "period-end", form.PeriodEnd),
		Amount: Money(sandbox, &reader, "amount", form.Amount), Notes: form.Notes, Months: Int(sandbox, &reader, "months", form.Months),
	}
	if fields.Amount < 0 {
		fields.Amount = -fields.Amount
	}
	return fields, ProblemOf(sandbox, reader)
}

// BudgetFormOf is the BudgetForm that fills the form with fields.
func BudgetFormOf(sandbox *api.Sandbox, lang string, fields erpstore.BudgetFields) BudgetForm {
	return BudgetForm{
		Company: id(sandbox, fields.Company), Category: id(sandbox, fields.Category),
		PeriodStart: erpdate.ISO(sandbox, fields.PeriodStart), PeriodEnd: erpdate.ISO(sandbox, fields.PeriodEnd),
		Amount: moneyInput(sandbox, lang, fields.Amount), Notes: fields.Notes,
	}
}

// BudgetFormPage is what erp/budget_form.html is rendered with.
type BudgetFormPage struct {
	Form       Form
	Values     BudgetForm
	Companies  []Option
	Categories []Option
	IsEdit     bool
}

// RenderBudgetForm answers the form adding budgets (id 0) or editing the one
// with id id, filled with values, problem above it.
func RenderBudgetForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, values BudgetForm, problem erpstore.Problem) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	if values.PeriodStart == "" {
		values.PeriodStart = erpdate.ISO(sandbox, erpdate.MonthStart(sandbox, book.Today))
		values.PeriodEnd = erpdate.ISO(sandbox, erpdate.MonthEnd(sandbox, book.Today))
	}
	form := newForm(sandbox, id, "budget", "budget", "/admin/erp/list-budgets", problem)
	body := BudgetFormPage{
		Form: form, Values: values, IsEdit: id != 0,
		Companies:  CompanyOptions(sandbox, book.Companies, ParseId(sandbox, values.Company)),
		Categories: CategoryOptions(sandbox, book.Categories, "", ParseId(sandbox, values.Category)),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/budget_form.html", "budgets", form, body)
}

// SettingsPage is what erp/settings.html is rendered with.
type SettingsPage struct {
	Currency    string
	Error       string
	UtcOffset   int64
	HorizonDays int64
	Today       int64
}

// RenderSettingsPage answers the settings page, filled with settings,
// problem above it.
func RenderSettingsPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, settings erpstore.Settings, problem erpstore.Problem, notice string) error {
	status := api.StatusOK
	if erpstore.Failed(sandbox, problem) {
		status = api.StatusBadRequest
	}
	body := SettingsPage{Currency: settings.Currency, Error: problem.Message, UtcOffset: sandbox.Config.UtcOffsetMinutes, HorizonDays: sandbox.Config.RecurrenceHorizonDays, Today: erpdate.Today(sandbox)}
	page := erppage.NewPage(sandbox, props, "Settings", "settings", NoticeOf(sandbox, notice), body)
	return erppage.Write(sandbox, response, status, "backoffice/erp/settings.html", page)
}
