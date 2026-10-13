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

// PageSize is how many transactions one page of the list shows.
const PageSize = 100

// TransactionFilters is what the filter form of the transactions sent, as
// it sent it, to fill the form again and to build the links of the list.
type TransactionFilters struct {
	Search           string
	Company          string
	Account          string
	Category         string
	ThirdParty       string
	Type             string
	Status           string
	From             string
	To               string
	DateField        string
	CardInvoice      string
	InstallmentGroup string
	Recurrence       string
	Page             int64
}

// QueryOf is the TransactionQuery filters ask for. With no period and nothing
// that names a set of its own — an installment group, a recurrence, an
// invoice — the period is this month.
func QueryOf(sandbox *api.Sandbox, filters *TransactionFilters) erpstore.TransactionQuery {
	reader := Reader{}
	query := erpstore.TransactionQuery{
		Search: filters.Search, Company: ParseId(sandbox, filters.Company), Account: ParseId(sandbox, filters.Account),
		Category: ParseId(sandbox, filters.Category), ThirdParty: ParseId(sandbox, filters.ThirdParty), Type: filters.Type,
		Status: filters.Status, From: Date(sandbox, &reader, "from", filters.From), To: Date(sandbox, &reader, "to", filters.To),
		DateField: filters.DateField, CardInvoice: Month(sandbox, &reader, "card-invoice", filters.CardInvoice),
		InstallmentGroup: ParseId(sandbox, filters.InstallmentGroup), Recurrence: ParseId(sandbox, filters.Recurrence),
	}
	if query.From == 0 && query.To == 0 && query.InstallmentGroup == 0 && query.Recurrence == 0 && query.CardInvoice == 0 && filters.From == "" && filters.To == "" {
		today := erpdate.Today(sandbox)
		query.From, query.To = erpdate.MonthStart(sandbox, today), erpdate.MonthEnd(sandbox, today)
		filters.From, filters.To = erpdate.ISO(sandbox, query.From), erpdate.ISO(sandbox, query.To)
	}
	return query
}

// SelfOf is the address of the list filters asks for, on page page.
func SelfOf(sandbox *api.Sandbox, filters TransactionFilters, page int64) string {
	return Href(sandbox, "/admin/erp/list-transactions",
		"search", filters.Search, "company", filters.Company, "account", filters.Account, "category", filters.Category,
		"third-party", filters.ThirdParty, "type", filters.Type, "status", filters.Status, "from", filters.From, "to", filters.To,
		"date-field", filters.DateField, "card-invoice", filters.CardInvoice, "installment-group", filters.InstallmentGroup,
		"recurrence", filters.Recurrence, "page", sandbox.Deps.StringsDeps.FormatInt(page, 10))
}

// TxRow is one transaction of a list.
type TxRow struct {
	Tx         erp_db.TransactionRecord
	Company    string
	Account    string
	Category   string
	ThirdParty string
	// Other names the account and company of the other leg of a transfer.
	Other      string
	IsTransfer bool
	IsCard     bool
	Overdue    bool
	// Installment spells "3/10" for the third of ten installments.
	Installment string
	// EditHref is where the transaction is edited: the transfer form for a
	// transfer leg.
	EditHref string
	// CanSettle and CanUnsettle offer the two actions on the row.
	CanSettle   bool
	CanUnsettle bool
}

// RowOf is the row of tx, named from book.
func RowOf(sandbox *api.Sandbox, book erpreports.Book, byId map[int64]erp_db.TransactionRecord, tx erp_db.TransactionRecord) TxRow {
	id := sandbox.Deps.StringsDeps.FormatInt(tx.Id, 10)
	row := TxRow{
		Tx: tx, Company: book.CompanyById[tx.Company].Name, Account: book.AccountById[tx.Account].Name,
		Category: book.CategoryPath[tx.Category], ThirdParty: book.ThirdPartyById[tx.ThirdParty].Name,
		IsTransfer: erpledger.IsTransfer(sandbox, tx.Type), IsCard: erpreports.IsCardAccount(sandbox, book, tx.Account),
		Overdue:  tx.Status == erpledger.StatusPending && erpledger.ExpectedDate(sandbox, tx) < book.Today,
		EditHref: "/admin/manage/set-transaction/" + id,
	}
	if tx.InstallmentCount > 1 {
		row.Installment = sandbox.Deps.StdDeps.Sprintf("%d/%d", tx.InstallmentNumber, tx.InstallmentCount)
	}
	if row.IsTransfer {
		row.EditHref = "/admin/manage/set-transfer/" + id
		if other, ok := byId[tx.Counterpart]; ok {
			row.Other = book.AccountById[other.Account].Name
			if other.Company != tx.Company {
				row.Other += " · " + book.CompanyById[other.Company].Name
			}
		}
	}
	charge := row.IsCard && erpledger.IsCharge(sandbox, tx)
	row.CanSettle = tx.Status == erpledger.StatusPending
	row.CanUnsettle = tx.Status == erpledger.StatusSettled && !charge
	return row
}

// Rows are the rows of txs, named from book.
func Rows(sandbox *api.Sandbox, book erpreports.Book, txs []erp_db.TransactionRecord) []TxRow {
	byId := map[int64]erp_db.TransactionRecord{}
	for _, tx := range book.Transactions {
		byId[tx.Id] = tx
	}
	rows := []TxRow{}
	for _, tx := range txs {
		rows = append(rows, RowOf(sandbox, book, byId, tx))
	}
	return rows
}

// TransactionsPage is what erp/transactions.html is rendered with.
type TransactionsPage struct {
	Filters      TransactionFilters
	Companies    []Option
	Accounts     []Option
	Categories   []Option
	ThirdParties []Option
	Types        []Option
	Statuses     []Option
	DateFields   []Option
	Rows         []TxRow
	Count        int
	Inflows      int64
	Outflows     int64
	Net          int64
	Pending      int64
	Page         int64
	Pages        int64
	PrevHref     string
	NextHref     string
	// Self is the address of this very list, which a form returns to.
	Self string
	// TodayISO fills the date of the settle buttons.
	TodayISO string
}

// RenderTransactionsPage answers the list of the transactions filters keeps,
// with the notice code notice above it.
func RenderTransactionsPage(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, filters TransactionFilters, notice string) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	query := QueryOf(sandbox, &filters)
	txs := erpstore.FilterTransactions(sandbox, book.Transactions, book.Categories, query)
	body := TransactionsPage{
		Filters:      filters,
		Companies:    CompanyOptions(sandbox, book.Companies, query.Company),
		Accounts:     AccountOptions(sandbox, book.Accounts, query.Account, true),
		Categories:   CategoryOptions(sandbox, book.Categories, "", query.Category),
		ThirdParties: ThirdPartyOptions(sandbox, book.ThirdParties, query.ThirdParty),
		Types:        Choices(sandbox, []string{erpledger.TypeIncome, erpledger.TypeExpense, "transfer", erpledger.TypeAdjustment}, filters.Type),
		Statuses:     Choices(sandbox, []string{erpledger.StatusPending, erpledger.StatusSettled, "overdue"}, filters.Status),
		DateFields:   []Option{Choice(sandbox, "date", "Entry date", filters.DateField), Choice(sandbox, "due", "Due date", filters.DateField), Choice(sandbox, "settled", "Settlement date", filters.DateField)},
		Count:        len(txs),
		TodayISO:     erpdate.ISO(sandbox, book.Today),
	}
	for _, tx := range txs {
		if tx.Status == erpledger.StatusPending {
			body.Pending += tx.Amount
		}
		if tx.Amount >= 0 {
			body.Inflows += tx.Amount
		} else {
			body.Outflows -= tx.Amount
		}
	}
	body.Net = body.Inflows - body.Outflows
	body.Pages = (int64(len(txs)) + PageSize - 1) / PageSize
	if body.Pages < 1 {
		body.Pages = 1
	}
	body.Page = filters.Page
	if body.Page < 1 {
		body.Page = 1
	}
	if body.Page > body.Pages {
		body.Page = body.Pages
	}
	start := (body.Page - 1) * PageSize
	end := start + PageSize
	if end > int64(len(txs)) {
		end = int64(len(txs))
	}
	body.Rows = Rows(sandbox, book, txs[start:end])
	body.Self = SelfOf(sandbox, filters, body.Page)
	if body.Page > 1 {
		body.PrevHref = SelfOf(sandbox, filters, body.Page-1)
	}
	if body.Page < body.Pages {
		body.NextHref = SelfOf(sandbox, filters, body.Page+1)
	}
	return erppage.Render(sandbox, response, api.StatusOK, props, "backoffice/erp/transactions.html", "Transactions", "transactions", NoticeOf(sandbox, notice), body)
}

// TransactionForm is what the transaction form posts, as it posts it.
type TransactionForm struct {
	Company        string
	Account        string
	Category       string
	ThirdParty     string
	Type           string
	Amount         string
	Description    string
	DocumentNumber string
	Notes          string
	Date           string
	DueDate        string
	Settled        string
	SettledDate    string
	CardInvoice    string
	Installments   string
	Next           string
}

// FieldsOf reads the TransactionFields form posted, the problem naming the
// first value it could not read.
func FieldsOf(sandbox *api.Sandbox, form TransactionForm) (erpstore.TransactionFields, erpstore.Problem) {
	reader := Reader{}
	fields := erpstore.TransactionFields{
		Company: ParseId(sandbox, form.Company), Account: ParseId(sandbox, form.Account), Category: ParseId(sandbox, form.Category),
		ThirdParty: ParseId(sandbox, form.ThirdParty), Type: form.Type, Amount: Money(sandbox, &reader, "amount", form.Amount),
		Description: form.Description, DocumentNumber: form.DocumentNumber, Notes: form.Notes,
		Date: Date(sandbox, &reader, "date", form.Date), DueDate: Date(sandbox, &reader, "due-date", form.DueDate),
		SettledDate: Date(sandbox, &reader, "settled-date", form.SettledDate), CardInvoice: Month(sandbox, &reader, "card-invoice", form.CardInvoice),
		Installments: Int(sandbox, &reader, "installments", form.Installments), Status: erpledger.StatusPending,
	}
	if form.Settled != "" {
		fields.Status = erpledger.StatusSettled
	}
	if fields.Type != erpledger.TypeAdjustment && fields.Amount < 0 {
		fields.Amount = -fields.Amount
	}
	return fields, ProblemOf(sandbox, reader)
}

// FormOf is the TransactionForm that fills the form with fields, in lang.
func FormOf(sandbox *api.Sandbox, lang string, fields erpstore.TransactionFields) TransactionForm {
	form := TransactionForm{
		Company: id(sandbox, fields.Company), Account: id(sandbox, fields.Account), Category: id(sandbox, fields.Category),
		ThirdParty: id(sandbox, fields.ThirdParty), Type: fields.Type, Amount: moneyInput(sandbox, lang, fields.Amount),
		Description: fields.Description, DocumentNumber: fields.DocumentNumber, Notes: fields.Notes,
		Date: erpdate.ISO(sandbox, fields.Date), DueDate: erpdate.ISO(sandbox, fields.DueDate), SettledDate: erpdate.ISO(sandbox, fields.SettledDate),
		CardInvoice: erpdate.ISOMonth(sandbox, fields.CardInvoice),
	}
	if fields.Status == erpledger.StatusSettled {
		form.Settled = "on"
	}
	if fields.Installments > 1 {
		form.Installments = sandbox.Deps.StringsDeps.FormatInt(fields.Installments, 10)
	}
	return form
}

// TransactionFormPage is what erp/transaction_form.html is rendered with.
type TransactionFormPage struct {
	Form         Form
	Values       TransactionForm
	Types        []Option
	Companies    []Option
	Accounts     []Option
	Categories   []Option
	ThirdParties []Option
	// Cards are the ids of the credit cards, which the form tells apart.
	Cards string
	// Installment is "3/10" when the transaction edited is an installment.
	Installment string
	// Recurrence names the recurrence the transaction edited came from.
	Recurrence string
	// Settled tells whether the box is ticked.
	Settled bool
	IsEdit  bool
}

// RenderTransactionForm answers the form adding a transaction (id 0) or
// editing the one with id id, filled with values, problem above it, returning
// to next.
func RenderTransactionForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, values TransactionForm, problem erpstore.Problem) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	values.Next = SafeNext(sandbox, values.Next, "/admin/erp/list-transactions")
	if values.Type == "" {
		values.Type = erpledger.TypeExpense
	}
	form := newForm(sandbox, id, "transaction", "transaction", values.Next, problem)
	body := TransactionFormPage{
		Form: form, Values: values, IsEdit: id != 0, Settled: values.Settled != "",
		Types:        Choices(sandbox, []string{erpledger.TypeExpense, erpledger.TypeIncome, erpledger.TypeAdjustment}, values.Type),
		Companies:    CompanyOptions(sandbox, book.Companies, ParseId(sandbox, values.Company)),
		Accounts:     AccountOptions(sandbox, book.Accounts, ParseId(sandbox, values.Account), true),
		Categories:   CategoryOptions(sandbox, book.Categories, "", ParseId(sandbox, values.Category)),
		ThirdParties: ThirdPartyOptions(sandbox, book.ThirdParties, ParseId(sandbox, values.ThirdParty)),
	}
	if id != 0 {
		if tx, ok := erpstore.FindTransaction(sandbox, id); ok {
			if tx.InstallmentCount > 1 {
				body.Installment = sandbox.Deps.StdDeps.Sprintf("%d/%d", tx.InstallmentNumber, tx.InstallmentCount)
			}
			if recurrence, ok := erpstore.FindRecurrence(sandbox, tx.Recurrence); ok {
				body.Recurrence = recurrence.Description
			}
		}
	}
	cards := []string{}
	for _, card := range erpreports.Cards(sandbox, book) {
		cards = append(cards, sandbox.Deps.StringsDeps.FormatInt(card.Id, 10))
	}
	body.Cards = sandbox.Deps.StringsDeps.Join(cards, ",")
	return renderForm(sandbox, response, props, "backoffice/erp/transaction_form.html", "transactions", form, body)
}

// TransferForm is what the transfer form posts, as it posts it.
type TransferForm struct {
	FromAccount    string
	ToAccount      string
	FromCompany    string
	ToCompany      string
	Amount         string
	Description    string
	DocumentNumber string
	Notes          string
	Date           string
	DueDate        string
	Settled        string
	SettledDate    string
	CardInvoice    string
	Next           string
}

// TransferFieldsOf reads the TransferFields form posted.
func TransferFieldsOf(sandbox *api.Sandbox, form TransferForm) (erpstore.TransferFields, erpstore.Problem) {
	reader := Reader{}
	fields := erpstore.TransferFields{
		FromAccount: ParseId(sandbox, form.FromAccount), ToAccount: ParseId(sandbox, form.ToAccount),
		FromCompany: ParseId(sandbox, form.FromCompany), ToCompany: ParseId(sandbox, form.ToCompany),
		Amount: Money(sandbox, &reader, "amount", form.Amount), Description: form.Description,
		DocumentNumber: form.DocumentNumber, Notes: form.Notes, Date: Date(sandbox, &reader, "date", form.Date),
		DueDate: Date(sandbox, &reader, "due-date", form.DueDate), SettledDate: Date(sandbox, &reader, "settled-date", form.SettledDate),
		CardInvoice: Month(sandbox, &reader, "card-invoice", form.CardInvoice), Status: erpledger.StatusPending,
	}
	if form.Settled != "" {
		fields.Status = erpledger.StatusSettled
	}
	return fields, ProblemOf(sandbox, reader)
}

// TransferFormOf is the TransferForm that fills the form with fields, in lang.
func TransferFormOf(sandbox *api.Sandbox, lang string, fields erpstore.TransferFields) TransferForm {
	form := TransferForm{
		FromAccount: id(sandbox, fields.FromAccount), ToAccount: id(sandbox, fields.ToAccount),
		FromCompany: id(sandbox, fields.FromCompany), ToCompany: id(sandbox, fields.ToCompany),
		Amount: moneyInput(sandbox, lang, fields.Amount), Description: fields.Description, DocumentNumber: fields.DocumentNumber,
		Notes: fields.Notes, Date: erpdate.ISO(sandbox, fields.Date), DueDate: erpdate.ISO(sandbox, fields.DueDate),
		SettledDate: erpdate.ISO(sandbox, fields.SettledDate), CardInvoice: erpdate.ISOMonth(sandbox, fields.CardInvoice),
	}
	if fields.Status == erpledger.StatusSettled {
		form.Settled = "on"
	}
	return form
}

// TransferFormPage is what erp/transfer_form.html is rendered with.
type TransferFormPage struct {
	Form          Form
	Values        TransferForm
	FromAccounts  []Option
	ToAccounts    []Option
	FromCompanies []Option
	ToCompanies   []Option
	Settled       bool
}

// RenderTransferForm answers the form adding a transfer (id 0) or editing
// the one a leg of which has id id, filled with values, problem above it.
func RenderTransferForm(sandbox *api.Sandbox, response *serverdeps.Response, props *routeprops.RouteProps, id int64, values TransferForm, problem erpstore.Problem) error {
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	values.Next = SafeNext(sandbox, values.Next, "/admin/erp/list-transactions")
	form := newForm(sandbox, id, "transfer", "transfer", values.Next, problem)
	body := TransferFormPage{
		Form: form, Values: values, Settled: values.Settled != "",
		FromAccounts:  AccountOptions(sandbox, book.Accounts, ParseId(sandbox, values.FromAccount), true),
		ToAccounts:    AccountOptions(sandbox, book.Accounts, ParseId(sandbox, values.ToAccount), true),
		FromCompanies: CompanyOptions(sandbox, book.Companies, ParseId(sandbox, values.FromCompany)),
		ToCompanies:   CompanyOptions(sandbox, book.Companies, ParseId(sandbox, values.ToCompany)),
	}
	if id != 0 {
		body.Form.RemoveAction = "/admin/manage/remove-transaction/" + sandbox.Deps.StringsDeps.FormatInt(id, 10)
	}
	return renderForm(sandbox, response, props, "backoffice/erp/transfer_form.html", "transactions", body.Form, body)
}
