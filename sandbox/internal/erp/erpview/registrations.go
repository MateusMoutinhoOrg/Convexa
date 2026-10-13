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

// The pages of the registrations — companies, accounts, categories and third
// parties — each a list and a form that adds or edits one.

// Form is what every form page shares.
type Form struct {
	// Title heads the form; Submit labels its button.
	Title  string
	Submit string
	// Action is where the form posts.
	Action string
	// Error is shown above the form, "" for none; ErrorField names the
	// field at fault.
	Error      string
	ErrorField string
	// Id is the record edited, 0 when adding one.
	Id int64
	// RemoveAction is where the remove button posts, "" when adding.
	RemoveAction string
	// Back is the list the form returns to.
	Back string
}

// newForm is the Form adding a record (id 0) or editing the one with id id.
func newForm(sandbox *api.Sandbox, id int64, noun string, slug string, back string, problem erpstore.Problem) Form {
	form := Form{Back: back, Error: problem.Message, ErrorField: problem.Field, Id: id}
	if id == 0 {
		form.Title = "Add " + noun
		form.Submit = "Add " + noun
		form.Action = "/admin/manage/add-" + slug
		return form
	}
	number := sandbox.Deps.StringsDeps.FormatInt(id, 10)
	form.Title = "Edit " + noun
	form.Submit = "Save changes"
	form.Action = "/admin/manage/set-" + slug + "/" + number
	form.RemoveAction = "/admin/manage/remove-" + slug + "/" + number
	return form
}

// renderForm answers the form page template titled by form, translated —
// under 400 when problem is one.
func renderForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, template string, active string, form Form, body any) error {
	status := api.StatusOK
	if form.Error != "" {
		status = api.StatusBadRequest
	}
	page := erppage.NewPage(sandbox, props, form.Title, active, erppage.Notice{}, body)
	return erppage.Write(sandbox, response, status, template, page)
}

// CompanyRow is one company of the list, with its cash and its card debt
// today.
type CompanyRow struct {
	Company erp_db.CompanyRecord
	Cash    int64
	Debt    int64
}

// CompaniesPage is what erp/companies.html is rendered with.
type CompaniesPage struct {
	Search   string
	Status   string
	Statuses []Option
	Rows     []CompanyRow
}

// RenderCompaniesPage answers the company list, narrowed by query, with the
// notice code notice above it.
func RenderCompaniesPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, query erpstore.CompanyQuery, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	companies, err := erpstore.Companies(sandbox, query)
	if err != nil {
		return err
	}
	totals := erpreports.TotalsOf(sandbox, erpreports.AccountBalances(sandbox, book, book.Today))
	body := CompaniesPage{Search: query.Search, Status: query.Status, Statuses: Choices(sandbox, Statuses(sandbox), query.Status)}
	for _, company := range companies {
		body.Rows = append(body.Rows, CompanyRow{Company: company, Cash: totals.CashByCompany[company.Id], Debt: totals.DebtByCompany[company.Id]})
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/companies.html", "Companies", "companies", NoticeOf(sandbox, notice), body)
}

// CompanyFormPage is what erp/company_form.html is rendered with.
type CompanyFormPage struct {
	Form     Form
	Fields   erpstore.CompanyFields
	Statuses []Option
}

// RenderCompanyForm answers the form adding a company (id 0) or editing the
// one with id id, filled with fields, problem above it.
func RenderCompanyForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, fields erpstore.CompanyFields, problem erpstore.Problem) error {
	body := CompanyFormPage{
		Form:     newForm(sandbox, id, "company", "company", "/admin/erp/list-companies", problem),
		Fields:   fields,
		Statuses: Choices(sandbox, Statuses(sandbox), erpstoreStatus(sandbox, fields.Status)),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/company_form.html", "companies", body.Form, body)
}

// erpstoreStatus is status, or active when it is "".
func erpstoreStatus(sandbox *api.Sandbox, status string) string {
	if status == "" {
		return erpstore.StatusActive
	}
	return status
}

// AccountRow is one account of the list, with its balance today and the
// share of each company.
type AccountRow struct {
	Account erp_db.AccountRecord
	Balance erpreports.AccountBalance
	Shares  []Share
	IsCard  bool
	// PaymentAccount is the name of the account that pays a card's invoices.
	PaymentAccount string
}

// Share is one company's part of an account's balance.
type Share struct {
	Company string
	Amount  int64
}

// AccountsPage is what erp/accounts.html is rendered with.
type AccountsPage struct {
	Search   string
	Kind     string
	Status   string
	Kinds    []Option
	Statuses []Option
	Rows     []AccountRow
	// Cash and Debt total the balances listed.
	Cash int64
	Debt int64
}

// SharesOf are the shares of balance, by company name, from book.
func SharesOf(sandbox *api.Sandbox, book erpreports.Book, balance erpreports.AccountBalance) []Share {
	shares := []Share{}
	for _, company := range book.Companies {
		if amount, ok := balance.ByCompany[company.Id]; ok {
			shares = append(shares, Share{Company: company.Name, Amount: amount})
		}
	}
	return shares
}

// RenderAccountsPage answers the account list, narrowed by query, with the
// notice code notice above it.
func RenderAccountsPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, query erpstore.AccountQuery, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	accounts, err := erpstore.Accounts(sandbox, query)
	if err != nil {
		return err
	}
	balances := map[int64]erpreports.AccountBalance{}
	for _, balance := range erpreports.AccountBalances(sandbox, book, book.Today) {
		balances[balance.Account.Id] = balance
	}
	body := AccountsPage{
		Search: query.Search, Kind: query.Kind, Status: query.Status,
		Kinds: Choices(sandbox, erpledger.Kinds(sandbox), query.Kind), Statuses: Choices(sandbox, Statuses(sandbox), query.Status),
	}
	for _, account := range accounts {
		balance := balances[account.Id]
		row := AccountRow{Account: account, Balance: balance, Shares: SharesOf(sandbox, book, balance), IsCard: erpledger.IsCard(sandbox, account)}
		if row.IsCard {
			row.PaymentAccount = book.AccountById[account.PaymentAccount].Name
			body.Debt -= balance.Total
		} else {
			body.Cash += balance.Total
		}
		body.Rows = append(body.Rows, row)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/accounts.html", "Accounts", "accounts", NoticeOf(sandbox, notice), body)
}

// AccountFormPage is what erp/account_form.html is rendered with.
type AccountFormPage struct {
	Form   Form
	Fields erpstore.AccountFields
	// The amounts and dates as the form's inputs are filled.
	CreditLimit    string
	OpeningBalance string
	OpeningDate    string
	Kinds          []Option
	Statuses       []Option
	PaymentOptions []Option
	CompanyOptions []Option
}

// RenderAccountForm answers the form adding an account (id 0) or editing
// the one with id id, filled with fields, problem above it.
func RenderAccountForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, fields erpstore.AccountFields, problem erpstore.Problem) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	lang := erppage.Language(sandbox, props)
	if fields.Kind == "" {
		fields.Kind = erpledger.KindChecking
	}
	payers := []erp_db.AccountRecord{}
	for _, account := range book.Accounts {
		if account.Id != id {
			payers = append(payers, account)
		}
	}
	opening := fields.OpeningDate
	if opening == 0 {
		opening = book.Today
	}
	body := AccountFormPage{
		Form:           newForm(sandbox, id, "account", "account", "/admin/erp/list-accounts", problem),
		Fields:         fields,
		CreditLimit:    moneyInput(sandbox, lang, fields.CreditLimit),
		OpeningBalance: moneyInput(sandbox, lang, fields.OpeningBalance),
		OpeningDate:    erpdate.ISO(sandbox, opening),
		Kinds:          Choices(sandbox, erpledger.Kinds(sandbox), fields.Kind),
		Statuses:       Choices(sandbox, Statuses(sandbox), erpstoreStatus(sandbox, fields.Status)),
		PaymentOptions: AccountOptions(sandbox, payers, fields.PaymentAccount, false),
		CompanyOptions: CompanyOptions(sandbox, book.Companies, fields.OpeningCompany),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/account_form.html", "accounts", body.Form, body)
}

// CategoryRow is one category of the list.
type CategoryRow struct {
	Category erp_db.CategoryRecord
	Path     string
	// Depth is how many categories it sits under.
	Depth int
	// Indent pads its name by its depth.
	Indent string
}

// CategoriesPage is what erp/categories.html is rendered with.
type CategoriesPage struct {
	Search   string
	Type     string
	Status   string
	Types    []Option
	Statuses []Option
	Rows     []CategoryRow
}

// RenderCategoriesPage answers the category list, narrowed by query, with
// the notice code notice above it.
func RenderCategoriesPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, query erpstore.CategoryQuery, notice string) error {
	all, err := erpstore.Categories(sandbox, erpstore.CategoryQuery{})
	if err != nil {
		return err
	}
	categories, err := erpstore.Categories(sandbox, query)
	if err != nil {
		return err
	}
	paths := erpstore.CategoryPaths(sandbox, all)
	body := CategoriesPage{
		Search: query.Search, Type: query.Type, Status: query.Status,
		Types:    Choices(sandbox, []string{erpledger.TypeIncome, erpledger.TypeExpense}, query.Type),
		Statuses: Choices(sandbox, Statuses(sandbox), query.Status),
	}
	for _, category := range categories {
		depth := sandbox.Deps.StringsDeps.Count(paths[category.Id], " › ")
		body.Rows = append(body.Rows, CategoryRow{Category: category, Path: paths[category.Id], Depth: depth, Indent: sandbox.Deps.StringsDeps.Repeat("— ", depth)})
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/categories.html", "Categories", "categories", NoticeOf(sandbox, notice), body)
}

// CategoryFormPage is what erp/category_form.html is rendered with.
type CategoryFormPage struct {
	Form     Form
	Fields   erpstore.CategoryFields
	Types    []Option
	Statuses []Option
	Parents  []Option
}

// RenderCategoryForm answers the form adding a category (id 0) or editing
// the one with id id, filled with fields, problem above it.
func RenderCategoryForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, fields erpstore.CategoryFields, problem erpstore.Problem) error {
	all, err := erpstore.Categories(sandbox, erpstore.CategoryQuery{})
	if err != nil {
		return err
	}
	if fields.Type == "" {
		fields.Type = erpledger.TypeExpense
	}
	parents := []erp_db.CategoryRecord{}
	excluded := map[int64]bool{}
	if id != 0 {
		excluded = erpstore.CategorySubtree(sandbox, all, id)
	}
	for _, category := range all {
		if !excluded[category.Id] {
			parents = append(parents, category)
		}
	}
	body := CategoryFormPage{
		Form:     newForm(sandbox, id, "category", "category", "/admin/erp/list-categories", problem),
		Fields:   fields,
		Types:    Choices(sandbox, []string{erpledger.TypeIncome, erpledger.TypeExpense}, fields.Type),
		Statuses: Choices(sandbox, Statuses(sandbox), erpstoreStatus(sandbox, fields.Status)),
		Parents:  CategoryOptions(sandbox, parents, "", fields.Parent),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/category_form.html", "categories", body.Form, body)
}

// ThirdPartiesPage is what erp/third_parties.html is rendered with.
type ThirdPartiesPage struct {
	Search   string
	Role     string
	Status   string
	Roles    []Option
	Statuses []Option
	Rows     []erp_db.ThirdPartyRecord
}

// Roles are the roles of a third party.
func Roles(sandbox *api.Sandbox) []string {
	return []string{erpstore.RoleCustomer, erpstore.RoleSupplier, erpstore.RoleBoth}
}

// RenderThirdPartiesPage answers the third party list, narrowed by query,
// with the notice code notice above it.
func RenderThirdPartiesPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, query erpstore.ThirdPartyQuery, notice string) error {
	parties, err := erpstore.ThirdParties(sandbox, query)
	if err != nil {
		return err
	}
	body := ThirdPartiesPage{
		Search: query.Search, Role: query.Role, Status: query.Status, Rows: parties,
		Roles:    Choices(sandbox, Roles(sandbox), query.Role),
		Statuses: Choices(sandbox, Statuses(sandbox), query.Status),
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/third_parties.html", "Third parties", "third-parties", NoticeOf(sandbox, notice), body)
}

// ThirdPartyFormPage is what erp/third_party_form.html is rendered with.
type ThirdPartyFormPage struct {
	Form     Form
	Fields   erpstore.ThirdPartyFields
	Roles    []Option
	Statuses []Option
}

// RenderThirdPartyForm answers the form adding a third party (id 0) or
// editing the one with id id, filled with fields, problem above it.
func RenderThirdPartyForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, fields erpstore.ThirdPartyFields, problem erpstore.Problem) error {
	if fields.Role == "" {
		fields.Role = erpstore.RoleBoth
	}
	body := ThirdPartyFormPage{
		Form:     newForm(sandbox, id, "third party", "third-party", "/admin/erp/list-third-parties", problem),
		Fields:   fields,
		Roles:    Choices(sandbox, Roles(sandbox), fields.Role),
		Statuses: Choices(sandbox, Statuses(sandbox), erpstoreStatus(sandbox, fields.Status)),
	}
	return renderForm(sandbox, response, props, "backoffice/erp/third_party_form.html", "third-parties", body.Form, body)
}
