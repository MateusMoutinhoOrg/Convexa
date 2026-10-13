package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// An account is where money is kept: a checking or a savings account, an
// investment, a wallet, a credit card. An account belongs to no company:
// several companies may move money through one, and each transaction says
// whose money it moves, so the account has one balance — the cash it holds —
// and a balance for each company.

// AccountFields is what an account is written from.
type AccountFields struct {
	Name        string
	Kind        string
	Institution string
	Notes       string
	Status      string
	// CreditLimit, ClosingDay, DueDay and PaymentAccount describe a credit
	// card, and are cleared for any other kind.
	CreditLimit    int64
	ClosingDay     int64
	DueDay         int64
	PaymentAccount int64
	// OpeningBalance, written with OpeningCompany on OpeningDate, adds the
	// adjustment that opens the account's balance, when a new account is
	// added with one other than 0.
	OpeningBalance int64
	OpeningCompany int64
	OpeningDate    int64
}

// AccountFieldsOf is what account was written from.
func AccountFieldsOf(sandbox *api.Sandbox, account erp_db.AccountRecord) AccountFields {
	return AccountFields{
		Name: account.Name, Kind: account.Kind, Institution: account.Institution, Notes: account.Notes, Status: account.Status,
		CreditLimit: account.CreditLimit, ClosingDay: account.ClosingDay, DueDay: account.DueDay, PaymentAccount: account.PaymentAccount,
	}
}

// AccountQuery narrows a listing of the accounts.
type AccountQuery struct {
	// Search keeps the accounts whose name or institution holds it.
	Search string
	// Kind keeps the accounts of one kind, "" for every one.
	Kind string
	// Status keeps the accounts of one status, "" for every one.
	Status string
	// Cards keeps only the credit cards when "only", and leaves them out
	// when "none"; "" keeps both.
	Cards string
}

// Accounts is every account query keeps, by name.
func Accounts(sandbox *api.Sandbox, query AccountQuery) ([]erp_db.AccountRecord, error) {
	all, err := db(sandbox).ListAccounts(erp_db.AccountFilter{})
	if err != nil {
		return nil, err
	}
	kept := []erp_db.AccountRecord{}
	for _, account := range all {
		if query.Kind != "" && account.Kind != query.Kind {
			continue
		}
		if query.Status != "" && account.Status != query.Status {
			continue
		}
		card := erpledger.IsCard(sandbox, account)
		if (query.Cards == "only" && !card) || (query.Cards == "none" && card) {
			continue
		}
		if !Matches(sandbox, account.Name+" "+account.Institution, query.Search) {
			continue
		}
		kept = append(kept, account)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool { return lower(sandbox, kept[i].Name) < lower(sandbox, kept[j].Name) })
	return kept, nil
}

// FindAccount is the account with id id.
func FindAccount(sandbox *api.Sandbox, id int64) (erp_db.AccountRecord, bool) {
	if id <= 0 {
		return erp_db.AccountRecord{}, false
	}
	return db(sandbox).FindAccountById(id)
}

// validateAccount cleans fields and tells what is wrong with them, for the
// account with id id (0 for a new one).
func validateAccount(sandbox *api.Sandbox, id int64, fields *AccountFields) (Problem, error) {
	fields.Name = clean(sandbox, fields.Name, 120)
	fields.Institution = clean(sandbox, fields.Institution, 120)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	fields.Status = statusOrActive(sandbox, fields.Status)
	if fields.Kind == "" {
		fields.Kind = erpledger.KindChecking
	}
	if fields.Name == "" {
		return invalid(sandbox, "name", "Enter the account's name."), nil
	}
	if !oneOf(sandbox, fields.Kind, erpledger.Kinds(sandbox)...) {
		return invalid(sandbox, "kind", "Choose a valid kind of account."), nil
	}
	if !oneOf(sandbox, fields.Status, StatusActive, StatusArchived) {
		return invalid(sandbox, "status", "The status must be active or archived."), nil
	}
	if fields.Kind == erpledger.KindCreditCard {
		if fields.ClosingDay < 1 || fields.ClosingDay > 31 {
			return invalid(sandbox, "closing-day", "The closing day must be between 1 and 31."), nil
		}
		if fields.DueDay < 1 || fields.DueDay > 31 {
			return invalid(sandbox, "due-day", "The due day must be between 1 and 31."), nil
		}
		if fields.CreditLimit < 0 {
			return invalid(sandbox, "credit-limit", "The credit limit cannot be negative."), nil
		}
		if fields.PaymentAccount != 0 {
			payer, ok := FindAccount(sandbox, fields.PaymentAccount)
			if !ok || erpledger.IsCard(sandbox, payer) || payer.Id == id {
				return invalid(sandbox, "payment-account", "The account that pays the invoices must be an existing account that is not a credit card."), nil
			}
		}
	} else {
		fields.CreditLimit, fields.ClosingDay, fields.DueDay, fields.PaymentAccount = 0, 0, 0, 0
	}
	all, err := db(sandbox).ListAccounts(erp_db.AccountFilter{})
	if err != nil {
		return Problem{}, err
	}
	for _, other := range all {
		if other.Id != id && sameText(sandbox, other.Name, fields.Name) {
			return invalid(sandbox, "name", "Another account already has that name."), nil
		}
	}
	if id != 0 {
		current, _ := FindAccount(sandbox, id)
		if (current.Kind == erpledger.KindCreditCard) != (fields.Kind == erpledger.KindCreditCard) {
			txs, err := Transactions(sandbox)
			if err != nil {
				return Problem{}, err
			}
			for _, tx := range txs {
				if tx.Account == id {
					return invalid(sandbox, "kind", "An account with transactions cannot turn into a credit card, or stop being one."), nil
				}
			}
		}
	}
	if id == 0 && fields.OpeningBalance != 0 {
		if _, ok := FindCompany(sandbox, fields.OpeningCompany); !ok {
			return invalid(sandbox, "opening-company", "Choose the company the opening balance belongs to."), nil
		}
		if fields.OpeningDate == 0 {
			fields.OpeningDate = erpdate.Today(sandbox)
		}
		if !erpdate.Valid(sandbox, fields.OpeningDate) {
			return invalid(sandbox, "opening-date", "Enter a valid date for the opening balance."), nil
		}
	}
	return Problem{}, nil
}

// AddAccount writes a new account from fields, with the adjustment that
// opens its balance when fields carries one.
func AddAccount(sandbox *api.Sandbox, fields AccountFields, userId int64) (erp_db.AccountRecord, Problem, error) {
	defer lock(sandbox)()
	problem, err := validateAccount(sandbox, 0, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.AccountRecord{}, problem, err
	}
	account, err := db(sandbox).AddAccount(erp_db.AccountInput{
		Name: fields.Name, Kind: fields.Kind, Institution: fields.Institution, Notes: fields.Notes, Status: fields.Status,
		CreditLimit: fields.CreditLimit, ClosingDay: fields.ClosingDay, DueDay: fields.DueDay, PaymentAccount: fields.PaymentAccount,
		CreatedAt: nowSeconds(sandbox),
	})
	if err != nil || fields.OpeningBalance == 0 {
		return account, Problem{}, err
	}
	opening := erp_db.TransactionInput{
		Company: fields.OpeningCompany, Account: account.Id, Type: erpledger.TypeAdjustment, Amount: fields.OpeningBalance,
		Description: "Opening balance", Date: fields.OpeningDate, DueDate: fields.OpeningDate, SettledDate: fields.OpeningDate,
		Status: erpledger.StatusSettled, CreatedAt: nowSeconds(sandbox), UpdatedAt: nowSeconds(sandbox), CreatedBy: userId,
	}
	if erpledger.IsCard(sandbox, account) {
		opening.CardInvoice = erpledger.InvoiceFor(sandbox, account, fields.OpeningDate)
	}
	_, err = db(sandbox).AddTransaction(opening)
	Invalidate(sandbox)
	return account, Problem{}, err
}

// SetAccount rewrites the account with id id from fields.
func SetAccount(sandbox *api.Sandbox, id int64, fields AccountFields) (erp_db.AccountRecord, Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindAccount(sandbox, id); !ok {
		return erp_db.AccountRecord{}, notFound(sandbox, "That account no longer exists."), nil
	}
	fields.OpeningBalance = 0
	problem, err := validateAccount(sandbox, id, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.AccountRecord{}, problem, err
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetAccountName(id, fields.Name),
		store.SetAccountKind(id, fields.Kind),
		store.SetAccountInstitution(id, fields.Institution),
		store.SetAccountNotes(id, fields.Notes),
		store.SetAccountStatus(id, fields.Status),
		store.SetAccountCreditLimit(id, fields.CreditLimit),
		store.SetAccountClosingDay(id, fields.ClosingDay),
		store.SetAccountDueDay(id, fields.DueDay),
		store.SetAccountPaymentAccount(id, fields.PaymentAccount),
	} {
		if write != nil {
			return erp_db.AccountRecord{}, Problem{}, write
		}
	}
	account, _ := FindAccount(sandbox, id)
	return account, Problem{}, nil
}

// RemoveAccount deletes the account with id id, refused while a transaction
// or a recurrence names it, or a card has it pay its invoices: archive it
// instead.
func RemoveAccount(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindAccount(sandbox, id); !ok {
		return notFound(sandbox, "That account no longer exists."), nil
	}
	txs, err := Transactions(sandbox)
	if err != nil {
		return Problem{}, err
	}
	for _, tx := range txs {
		if tx.Account == id {
			return conflict(sandbox, "id", "This account has transactions, so it cannot be removed. Archive it instead."), nil
		}
	}
	recurrences, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{AccountMin: id, AccountMax: id})
	if err != nil {
		return Problem{}, err
	}
	if len(recurrences) > 0 {
		return conflict(sandbox, "id", "A recurrence uses this account, so it cannot be removed. Archive it instead."), nil
	}
	cards, err := db(sandbox).ListAccounts(erp_db.AccountFilter{PaymentAccountMin: id, PaymentAccountMax: id})
	if err != nil {
		return Problem{}, err
	}
	if len(cards) > 0 {
		return conflict(sandbox, "id", "A credit card pays its invoices from this account, so it cannot be removed."), nil
	}
	return Problem{}, db(sandbox).RemoveAccount(id)
}
