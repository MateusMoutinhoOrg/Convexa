package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// The reports of Convexa are computed over one Book: every record read once,
// at the start of a request, indexed by id. A report reads nothing else, so
// the numbers of one page always agree with each other.

// Book is every financial record, read at once.
type Book struct {
	Today        int64
	Currency     string
	Companies    []erp_db.CompanyRecord
	Accounts     []erp_db.AccountRecord
	Categories   []erp_db.CategoryRecord
	ThirdParties []erp_db.ThirdPartyRecord
	Transactions []erp_db.TransactionRecord
	Recurrences  []erp_db.RecurrenceRecord
	Budgets      []erp_db.BudgetRecord

	CompanyById    map[int64]erp_db.CompanyRecord
	AccountById    map[int64]erp_db.AccountRecord
	CategoryById   map[int64]erp_db.CategoryRecord
	ThirdPartyById map[int64]erp_db.ThirdPartyRecord
	// CategoryPath is the full name of every category: "Housing › Rent".
	CategoryPath map[int64]string
}

// Load reads the Book.
func Load(sandbox *api.Sandbox) (Book, error) {
	book := Book{Today: erpdate.Today(sandbox), Currency: erpstore.Currency(sandbox)}
	var err error
	if book.Companies, err = erpstore.Companies(sandbox, erpstore.CompanyQuery{}); err != nil {
		return book, err
	}
	if book.Accounts, err = erpstore.Accounts(sandbox, erpstore.AccountQuery{}); err != nil {
		return book, err
	}
	if book.Categories, err = erpstore.Categories(sandbox, erpstore.CategoryQuery{}); err != nil {
		return book, err
	}
	if book.ThirdParties, err = erpstore.ThirdParties(sandbox, erpstore.ThirdPartyQuery{}); err != nil {
		return book, err
	}
	if book.Transactions, err = erpstore.Transactions(sandbox); err != nil {
		return book, err
	}
	if book.Recurrences, err = erpstore.Recurrences(sandbox, erpstore.RecurrenceQuery{}); err != nil {
		return book, err
	}
	if book.Budgets, err = erpstore.Budgets(sandbox, erpstore.BudgetQuery{}); err != nil {
		return book, err
	}
	book.CompanyById = map[int64]erp_db.CompanyRecord{}
	for _, company := range book.Companies {
		book.CompanyById[company.Id] = company
	}
	book.AccountById = map[int64]erp_db.AccountRecord{}
	for _, account := range book.Accounts {
		book.AccountById[account.Id] = account
	}
	book.CategoryById = map[int64]erp_db.CategoryRecord{}
	for _, category := range book.Categories {
		book.CategoryById[category.Id] = category
	}
	book.ThirdPartyById = map[int64]erp_db.ThirdPartyRecord{}
	for _, party := range book.ThirdParties {
		book.ThirdPartyById[party.Id] = party
	}
	book.CategoryPath = erpstore.CategoryPaths(sandbox, book.Categories)
	return book, nil
}

// IsCardAccount tells whether the account with id id is a credit card.
func IsCardAccount(sandbox *api.Sandbox, book Book, id int64) bool {
	account, ok := book.AccountById[id]
	return ok && erpledger.IsCard(sandbox, account)
}

// CashAccounts are the accounts that hold cash: every one but the credit
// cards.
func CashAccounts(sandbox *api.Sandbox, book Book) []erp_db.AccountRecord {
	cash := []erp_db.AccountRecord{}
	for _, account := range book.Accounts {
		if !erpledger.IsCard(sandbox, account) {
			cash = append(cash, account)
		}
	}
	return cash
}

// Cards are the credit cards.
func Cards(sandbox *api.Sandbox, book Book) []erp_db.AccountRecord {
	cards := []erp_db.AccountRecord{}
	for _, account := range book.Accounts {
		if erpledger.IsCard(sandbox, account) {
			cards = append(cards, account)
		}
	}
	return cards
}

// Scope is the accounts and the company a report is narrowed to.
type Scope struct {
	// Accounts are the ids of the accounts in scope.
	Accounts map[int64]bool
	// Company is the id of the company in scope, 0 for every one.
	Company int64
	// AllCash is true when every cash account is in scope: what belongs to
	// no account — a budget, a card invoice with no payment account — is in
	// scope then.
	AllCash bool
}

// CashScope is the scope of the cash accounts — or of the account with id
// account alone, when not 0 — and the company with id company.
func CashScope(sandbox *api.Sandbox, book Book, account int64, company int64) Scope {
	scope := Scope{Accounts: map[int64]bool{}, Company: company}
	if account != 0 {
		scope.Accounts[account] = true
		return scope
	}
	for _, cash := range CashAccounts(sandbox, book) {
		scope.Accounts[cash.Id] = true
	}
	scope.AllCash = true
	return scope
}

// In tells whether a record of the account with id account and the company
// with id company is in scope.
func In(sandbox *api.Sandbox, scope Scope, account int64, company int64) bool {
	return scope.Accounts[account] && (scope.Company == 0 || scope.Company == company)
}

// counterparts is every transaction of book by id, to find the other leg of a
// transfer.
func counterparts(sandbox *api.Sandbox, book Book) map[int64]erp_db.TransactionRecord {
	byId := map[int64]erp_db.TransactionRecord{}
	for _, tx := range book.Transactions {
		byId[tx.Id] = tx
	}
	return byId
}

// Internal tells whether tx is a transfer leg whose other leg is in scope as
// well: a movement inside the scope, which changes none of its totals and is
// left out of its inflows and outflows.
func Internal(sandbox *api.Sandbox, scope Scope, byId map[int64]erp_db.TransactionRecord, tx erp_db.TransactionRecord) bool {
	if !erpledger.IsTransfer(sandbox, tx.Type) {
		return false
	}
	other, ok := byId[tx.Counterpart]
	return ok && In(sandbox, scope, other.Account, other.Company)
}

// Balance is the settled balance of scope at date.
func Balance(sandbox *api.Sandbox, book Book, scope Scope, date int64) int64 {
	total := int64(0)
	for _, tx := range book.Transactions {
		if In(sandbox, scope, tx.Account, tx.Company) && erpledger.SettledBy(sandbox, tx, date) {
			total += tx.Amount
		}
	}
	return total
}
