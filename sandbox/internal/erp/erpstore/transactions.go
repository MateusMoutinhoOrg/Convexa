package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// A transaction (lançamento) is one movement of money of a company on an
// account: an income, an expense, an adjustment, or one leg of a transfer. It
// carries the date it was entered for, the date it is due and — once it is
// paid or received — the date it settled. A pending one is a payable or a
// receivable; a settled one moved the balance.
//
// On a credit card, a purchase is settled the day it is made — the card owes
// it from then on, every installment of it — and lands on the invoice its date
// falls in; what is pending there is the payment of the invoice, a transfer
// from another account.

// MaxAmount is the largest amount a transaction carries, in cents.
const MaxAmount = int64(1_000_000_000_000_00)

// MaxInstallments is the most installments one purchase is split into.
const MaxInstallments = 360

// TransactionFields is what a transaction — not a transfer — is written from.
type TransactionFields struct {
	Company    int64
	Account    int64
	Category   int64
	ThirdParty int64
	// Type is income, expense or adjustment.
	Type string
	// Amount is in cents: positive for an income or an expense, whose type
	// gives its sign; signed for an adjustment.
	Amount         int64
	Description    string
	DocumentNumber string
	Notes          string
	Date           int64
	DueDate        int64
	SettledDate    int64
	// Status is pending or settled.
	Status string
	// CardInvoice is the invoice (YYYYMM) a transaction of a credit card
	// lands on; 0 works it out from the date.
	CardInvoice int64
	// Installments splits a new transaction into that many, a month apart;
	// 0 or 1 adds one.
	Installments int64
}

// TransactionFieldsOf is what tx was written from, its amount unsigned unless
// tx is an adjustment.
func TransactionFieldsOf(sandbox *api.Sandbox, tx erp_db.TransactionRecord) TransactionFields {
	amount := tx.Amount
	if tx.Type != erpledger.TypeAdjustment && amount < 0 {
		amount = -amount
	}
	return TransactionFields{
		Company: tx.Company, Account: tx.Account, Category: tx.Category, ThirdParty: tx.ThirdParty,
		Type: tx.Type, Amount: amount, Description: tx.Description, DocumentNumber: tx.DocumentNumber, Notes: tx.Notes,
		Date: tx.Date, DueDate: tx.DueDate, SettledDate: tx.SettledDate, Status: tx.Status, CardInvoice: tx.CardInvoice,
	}
}

// TransactionQuery narrows a listing of the transactions.
type TransactionQuery struct {
	// Search keeps the transactions whose description, document number or
	// notes hold it.
	Search  string
	Company int64
	Account int64
	// Category keeps the transactions of a category or of any category under
	// it.
	Category   int64
	ThirdParty int64
	// Type keeps one type; "transfer" keeps both legs of the transfers.
	Type string
	// Status keeps pending or settled ones; "overdue" keeps the pending ones
	// due before today.
	Status string
	// From and To keep the dates in between, both included; 0 is open.
	From int64
	To   int64
	// DateField is the date From and To read: "date" (the default), "due"
	// or "settled".
	DateField        string
	CardInvoice      int64
	InstallmentGroup int64
	Recurrence       int64
}

// dateOf is the date of tx that field names.
func dateOf(sandbox *api.Sandbox, tx erp_db.TransactionRecord, field string) int64 {
	switch field {
	case "due":
		if tx.DueDate != 0 {
			return tx.DueDate
		}
		return tx.Date
	case "settled":
		return tx.SettledDate
	}
	return tx.Date
}

// FilterTransactions is every transaction of txs query keeps, newest first.
// categories are every category, read to follow the subcategories.
func FilterTransactions(sandbox *api.Sandbox, txs []erp_db.TransactionRecord, categories []erp_db.CategoryRecord, query TransactionQuery) []erp_db.TransactionRecord {
	var subtree map[int64]bool
	if query.Category != 0 {
		subtree = CategorySubtree(sandbox, categories, query.Category)
	}
	today := erpdate.Today(sandbox)
	kept := []erp_db.TransactionRecord{}
	for _, tx := range txs {
		if query.Company != 0 && tx.Company != query.Company {
			continue
		}
		if query.Account != 0 && tx.Account != query.Account {
			continue
		}
		if subtree != nil && !subtree[tx.Category] {
			continue
		}
		if query.ThirdParty != 0 && tx.ThirdParty != query.ThirdParty {
			continue
		}
		if query.Type == "transfer" {
			if !erpledger.IsTransfer(sandbox, tx.Type) {
				continue
			}
		} else if query.Type != "" && tx.Type != query.Type {
			continue
		}
		switch query.Status {
		case erpledger.StatusPending, erpledger.StatusSettled:
			if tx.Status != query.Status {
				continue
			}
		case "overdue":
			if tx.Status != erpledger.StatusPending || erpledger.ExpectedDate(sandbox, tx) >= today {
				continue
			}
		}
		date := dateOf(sandbox, tx, query.DateField)
		if (query.From != 0 && date < query.From) || (query.To != 0 && date > query.To) {
			continue
		}
		if query.CardInvoice != 0 && tx.CardInvoice != query.CardInvoice {
			continue
		}
		if query.InstallmentGroup != 0 && tx.InstallmentGroup != query.InstallmentGroup {
			continue
		}
		if query.Recurrence != 0 && tx.Recurrence != query.Recurrence {
			continue
		}
		if !Matches(sandbox, tx.Description+" "+tx.DocumentNumber+" "+tx.Notes, query.Search) {
			continue
		}
		kept = append(kept, tx)
	}
	SortNewestFirst(sandbox, kept, query.DateField)
	return kept
}

// SortNewestFirst sorts txs by the date field names, newest first, the last
// written first among the ones of one day.
func SortNewestFirst(sandbox *api.Sandbox, txs []erp_db.TransactionRecord, field string) {
	sandbox.Deps.SortDeps.Slice(txs, func(i int, j int) bool {
		a, b := dateOf(sandbox, txs[i], field), dateOf(sandbox, txs[j], field)
		if a != b {
			return a > b
		}
		return txs[i].Id > txs[j].Id
	})
}

// ListTransactions is every transaction query keeps, newest first.
func ListTransactions(sandbox *api.Sandbox, query TransactionQuery) ([]erp_db.TransactionRecord, error) {
	txs, err := Transactions(sandbox)
	if err != nil {
		return nil, err
	}
	var categories []erp_db.CategoryRecord
	if query.Category != 0 {
		categories, err = db(sandbox).ListCategories(erp_db.CategoryFilter{})
		if err != nil {
			return nil, err
		}
	}
	return FilterTransactions(sandbox, txs, categories, query), nil
}

// FindTransaction is the transaction with id id.
func FindTransaction(sandbox *api.Sandbox, id int64) (erp_db.TransactionRecord, bool) {
	if id <= 0 {
		return erp_db.TransactionRecord{}, false
	}
	return db(sandbox).FindTransactionById(id)
}

// signed is amount with the sign a transaction of type kind carries.
func signed(sandbox *api.Sandbox, kind string, amount int64) int64 {
	switch kind {
	case erpledger.TypeExpense, erpledger.TypeTransferOut:
		return -amount
	}
	return amount
}

// defaultSettledDate is the date a transaction due on due settles on when it
// is marked settled with no date: its due date when that has come, today
// when it has not — what was paid ahead of time was paid today.
func defaultSettledDate(sandbox *api.Sandbox, due int64) int64 {
	today := erpdate.Today(sandbox)
	if due != 0 && due <= today {
		return due
	}
	return today
}

// checkCompany tells what is wrong with the company with id id for a record
// that named current before (0 for a new one): it must exist, and be active
// unless the record already named it.
func checkCompany(sandbox *api.Sandbox, field string, id int64, current int64) Problem {
	company, ok := FindCompany(sandbox, id)
	if !ok {
		return invalid(sandbox, field, "Choose the company.")
	}
	if company.Status != StatusActive && id != current {
		return invalid(sandbox, field, "That company is archived.")
	}
	return Problem{}
}

// checkAccount is the account with id id and what is wrong with it for a
// record that named current before (0 for a new one).
func checkAccount(sandbox *api.Sandbox, field string, id int64, current int64) (erp_db.AccountRecord, Problem) {
	account, ok := FindAccount(sandbox, id)
	if !ok {
		return account, invalid(sandbox, field, "Choose the account.")
	}
	if account.Status != StatusActive && id != current {
		return account, invalid(sandbox, field, "That account is archived.")
	}
	return account, Problem{}
}

// checkClassification tells what is wrong with category and third party for
// a transaction of type kind.
func checkClassification(sandbox *api.Sandbox, kind string, category int64, party int64) Problem {
	if category != 0 {
		found, ok := FindCategory(sandbox, category)
		if !ok {
			return invalid(sandbox, "category", "The category does not exist.")
		}
		if kind != erpledger.TypeIncome && kind != erpledger.TypeExpense {
			return invalid(sandbox, "category", "Only an income or an expense has a category.")
		}
		if found.Type != kind {
			return invalid(sandbox, "category", "The category is not of the transaction's type.")
		}
	}
	if party != 0 {
		if _, ok := FindThirdParty(sandbox, party); !ok {
			return invalid(sandbox, "third-party", "The third party does not exist.")
		}
	}
	return Problem{}
}

// checkDates cleans the dates and the status of a transaction and tells what
// is wrong with them.
func checkDates(sandbox *api.Sandbox, date int64, due *int64, status *string, settled *int64) Problem {
	if date == 0 || !erpdate.Valid(sandbox, date) {
		return invalid(sandbox, "date", "Enter a valid date.")
	}
	if *due == 0 {
		*due = date
	}
	if !erpdate.Valid(sandbox, *due) {
		return invalid(sandbox, "due-date", "Enter a valid due date.")
	}
	if *status == "" {
		*status = erpledger.StatusPending
	}
	switch *status {
	case erpledger.StatusPending:
		*settled = 0
	case erpledger.StatusSettled:
		if *settled == 0 {
			*settled = defaultSettledDate(sandbox, *due)
		}
		if !erpdate.Valid(sandbox, *settled) {
			return invalid(sandbox, "settled-date", "Enter a valid settlement date.")
		}
	default:
		return invalid(sandbox, "status", "The status must be pending or settled.")
	}
	return Problem{}
}

// validateTransaction cleans fields and tells what is wrong with them, for a
// transaction that was current before (nil for a new one), answering the
// account it lands on.
func validateTransaction(sandbox *api.Sandbox, fields *TransactionFields, current *erp_db.TransactionRecord) (erp_db.AccountRecord, Problem) {
	fields.Description = clean(sandbox, fields.Description, 200)
	fields.DocumentNumber = clean(sandbox, fields.DocumentNumber, 60)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	var currentCompany, currentAccount int64
	if current != nil {
		currentCompany, currentAccount = current.Company, current.Account
	}
	if !oneOf(sandbox, fields.Type, erpledger.TypeIncome, erpledger.TypeExpense, erpledger.TypeAdjustment) {
		return erp_db.AccountRecord{}, invalid(sandbox, "type", "The type must be income, expense or adjustment. A transfer is written with the transfer form.")
	}
	if problem := checkCompany(sandbox, "company", fields.Company, currentCompany); Failed(sandbox, problem) {
		return erp_db.AccountRecord{}, problem
	}
	account, problem := checkAccount(sandbox, "account", fields.Account, currentAccount)
	if Failed(sandbox, problem) {
		return account, problem
	}
	if problem := checkClassification(sandbox, fields.Type, fields.Category, fields.ThirdParty); Failed(sandbox, problem) {
		return account, problem
	}
	if fields.Type == erpledger.TypeAdjustment {
		if fields.Amount == 0 {
			return account, invalid(sandbox, "amount", "Enter the amount the balance is adjusted by: positive adds, negative subtracts.")
		}
		if fields.Description == "" {
			fields.Description = "Balance adjustment"
		}
	} else if fields.Amount <= 0 {
		return account, invalid(sandbox, "amount", "Enter an amount greater than zero.")
	}
	if fields.Amount > MaxAmount || fields.Amount < -MaxAmount {
		return account, invalid(sandbox, "amount", "The amount is too large.")
	}
	if fields.Description == "" {
		return account, invalid(sandbox, "description", "Enter a description.")
	}
	if erpledger.IsCard(sandbox, account) {
		fields.Status = erpledger.StatusSettled
		if fields.SettledDate == 0 {
			fields.SettledDate = fields.Date
		}
		fields.DueDate = fields.Date
	}
	if problem := checkDates(sandbox, fields.Date, &fields.DueDate, &fields.Status, &fields.SettledDate); Failed(sandbox, problem) {
		return account, problem
	}
	if erpledger.IsCard(sandbox, account) {
		if fields.CardInvoice == 0 {
			fields.CardInvoice = erpledger.InvoiceFor(sandbox, account, fields.Date)
		}
		if !erpdate.Valid(sandbox, fields.CardInvoice*100+1) {
			return account, invalid(sandbox, "card-invoice", "Enter the invoice as a valid month.")
		}
	} else {
		fields.CardInvoice = 0
	}
	if fields.Installments < 0 || fields.Installments > MaxInstallments {
		return account, invalid(sandbox, "installments", "The installments must be between 1 and 360.")
	}
	return account, Problem{}
}

// AddTransaction writes a new transaction from fields as the user with id
// userId — or, with fields.Installments above 1, that many, a month apart,
// the amount split between them and any cent left over on the first. The
// installments share an installment group, the id of the first.
func AddTransaction(sandbox *api.Sandbox, fields TransactionFields, userId int64) ([]erp_db.TransactionRecord, Problem, error) {
	defer lock(sandbox)()
	account, problem := validateTransaction(sandbox, &fields, nil)
	if Failed(sandbox, problem) {
		return nil, problem, nil
	}
	defer Invalidate(sandbox)
	count := fields.Installments
	if count < 1 {
		count = 1
	}
	if count > 1 && fields.Type == erpledger.TypeAdjustment {
		return nil, invalid(sandbox, "installments", "An adjustment is never split into installments."), nil
	}
	share := fields.Amount / count
	first := fields.Amount - share*(count-1)
	now := nowSeconds(sandbox)
	created := []erp_db.TransactionRecord{}
	group := int64(0)
	for index := int64(0); index < count; index++ {
		amount := share
		if index == 0 {
			amount = first
		}
		input := erp_db.TransactionInput{
			Company: fields.Company, Account: fields.Account, Category: fields.Category, ThirdParty: fields.ThirdParty,
			Type: fields.Type, Amount: signed(sandbox, fields.Type, amount), Description: fields.Description,
			DocumentNumber: fields.DocumentNumber, Notes: fields.Notes,
			Date: erpdate.AddMonths(sandbox, fields.Date, index), DueDate: erpdate.AddMonths(sandbox, fields.DueDate, index),
			Status: fields.Status, SettledDate: fields.SettledDate, CardInvoice: fields.CardInvoice,
			CreatedAt: now, UpdatedAt: now, CreatedBy: userId,
		}
		if count > 1 {
			input.Description = sandbox.Deps.StdDeps.Sprintf("%s (%d/%d)", fields.Description, index+1, count)
			input.InstallmentGroup = group
			input.InstallmentNumber = index + 1
			input.InstallmentCount = count
			if index > 0 {
				input.Status = erpledger.StatusPending
				input.SettledDate = 0
			}
		}
		if erpledger.IsCard(sandbox, account) {
			// The card owes the whole purchase from the day it is made — every
			// installment takes its limit — so each one is settled that day;
			// its date stays in its own month, where the reports and the
			// budgets count it, and it lands on that month's invoice.
			input.Status = erpledger.StatusSettled
			input.SettledDate = fields.Date
			input.DueDate = input.Date
			input.CardInvoice = erpdate.AddToMonth(sandbox, fields.CardInvoice, index)
		}
		tx, err := db(sandbox).AddTransaction(input)
		if err != nil {
			removeAll(sandbox, created)
			return nil, Problem{}, err
		}
		if count > 1 && index == 0 {
			group = tx.Id
			if err := db(sandbox).SetTransactionInstallmentGroup(tx.Id, group); err != nil {
				removeAll(sandbox, append(created, tx))
				return nil, Problem{}, err
			}
			tx.InstallmentGroup = group
		}
		created = append(created, tx)
	}
	return created, Problem{}, nil
}

// removeAll deletes every transaction of txs, as far as it can: it undoes a
// write of several that failed halfway.
func removeAll(sandbox *api.Sandbox, txs []erp_db.TransactionRecord) {
	for _, tx := range txs {
		_ = db(sandbox).RemoveTransaction(tx.Id)
	}
}

// SetTransaction rewrites the transaction with id id — not a transfer leg,
// which SetTransfer rewrites — from fields, as the user with id userId.
func SetTransaction(sandbox *api.Sandbox, id int64, fields TransactionFields, userId int64) (erp_db.TransactionRecord, Problem, error) {
	defer lock(sandbox)()
	current, ok := FindTransaction(sandbox, id)
	if !ok {
		return erp_db.TransactionRecord{}, notFound(sandbox, "That transaction no longer exists."), nil
	}
	if erpledger.IsTransfer(sandbox, current.Type) {
		return erp_db.TransactionRecord{}, invalid(sandbox, "type", "A transfer is edited with the transfer form."), nil
	}
	fields.Installments = 0
	_, problem := validateTransaction(sandbox, &fields, &current)
	if Failed(sandbox, problem) {
		return erp_db.TransactionRecord{}, problem, nil
	}
	defer Invalidate(sandbox)
	store := db(sandbox)
	for _, write := range []error{
		store.SetTransactionCompany(id, fields.Company),
		store.SetTransactionAccount(id, fields.Account),
		store.SetTransactionCategory(id, fields.Category),
		store.SetTransactionThirdParty(id, fields.ThirdParty),
		store.SetTransactionType(id, fields.Type),
		store.SetTransactionAmount(id, signed(sandbox, fields.Type, fields.Amount)),
		store.SetTransactionDescription(id, fields.Description),
		store.SetTransactionDocumentNumber(id, fields.DocumentNumber),
		store.SetTransactionNotes(id, fields.Notes),
		store.SetTransactionDate(id, fields.Date),
		store.SetTransactionDueDate(id, fields.DueDate),
		store.SetTransactionSettledDate(id, fields.SettledDate),
		store.SetTransactionStatus(id, fields.Status),
		store.SetTransactionCardInvoice(id, fields.CardInvoice),
		store.SetTransactionUpdatedAt(id, nowSeconds(sandbox)),
	} {
		if write != nil {
			return erp_db.TransactionRecord{}, Problem{}, write
		}
	}
	tx, _ := FindTransaction(sandbox, id)
	return tx, Problem{}, nil
}

// RemoveTransaction deletes the transaction with id id — both legs, for a
// transfer.
func RemoveTransaction(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	tx, ok := FindTransaction(sandbox, id)
	if !ok {
		return notFound(sandbox, "That transaction no longer exists."), nil
	}
	defer Invalidate(sandbox)
	if tx.Counterpart != 0 {
		if _, ok := FindTransaction(sandbox, tx.Counterpart); ok {
			if err := db(sandbox).RemoveTransaction(tx.Counterpart); err != nil {
				return Problem{}, err
			}
		}
	}
	return Problem{}, db(sandbox).RemoveTransaction(id)
}

// legsOf is the transaction tx and its counterpart, for a transfer leg; tx
// alone otherwise.
func legsOf(sandbox *api.Sandbox, tx erp_db.TransactionRecord) []erp_db.TransactionRecord {
	legs := []erp_db.TransactionRecord{tx}
	if tx.Counterpart != 0 {
		if other, ok := FindTransaction(sandbox, tx.Counterpart); ok {
			legs = append(legs, other)
		}
	}
	return legs
}

// SettleTransaction marks the transaction with id id — both legs, for a
// transfer — paid or received on date; today when date is 0.
func SettleTransaction(sandbox *api.Sandbox, id int64, date int64) (Problem, error) {
	defer lock(sandbox)()
	tx, ok := FindTransaction(sandbox, id)
	if !ok {
		return notFound(sandbox, "That transaction no longer exists."), nil
	}
	if date == 0 {
		date = erpdate.Today(sandbox)
	}
	if !erpdate.Valid(sandbox, date) {
		return invalid(sandbox, "settled-date", "Enter a valid settlement date."), nil
	}
	defer Invalidate(sandbox)
	for _, leg := range legsOf(sandbox, tx) {
		if err := db(sandbox).SetTransactionStatus(leg.Id, erpledger.StatusSettled); err != nil {
			return Problem{}, err
		}
		if err := db(sandbox).SetTransactionSettledDate(leg.Id, date); err != nil {
			return Problem{}, err
		}
		if err := db(sandbox).SetTransactionUpdatedAt(leg.Id, nowSeconds(sandbox)); err != nil {
			return Problem{}, err
		}
	}
	return Problem{}, nil
}

// UnsettleTransaction marks the transaction with id id — both legs, for a
// transfer — pending again. A purchase on a credit card is always settled:
// what is pending there is the payment of its invoice.
func UnsettleTransaction(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	tx, ok := FindTransaction(sandbox, id)
	if !ok {
		return notFound(sandbox, "That transaction no longer exists."), nil
	}
	legs := legsOf(sandbox, tx)
	for _, leg := range legs {
		account, _ := FindAccount(sandbox, leg.Account)
		if erpledger.IsCard(sandbox, account) && erpledger.IsCharge(sandbox, leg) {
			return conflict(sandbox, "status", "A credit card purchase is always settled: pay its invoice instead."), nil
		}
	}
	defer Invalidate(sandbox)
	for _, leg := range legs {
		if err := db(sandbox).SetTransactionStatus(leg.Id, erpledger.StatusPending); err != nil {
			return Problem{}, err
		}
		if err := db(sandbox).SetTransactionSettledDate(leg.Id, 0); err != nil {
			return Problem{}, err
		}
		if err := db(sandbox).SetTransactionUpdatedAt(leg.Id, nowSeconds(sandbox)); err != nil {
			return Problem{}, err
		}
	}
	return Problem{}, nil
}
