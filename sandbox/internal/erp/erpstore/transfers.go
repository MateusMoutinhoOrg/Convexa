package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// A transfer moves money from one account to another — or, on the same
// account, from one company to another — as two transactions pointing at each
// other: a transfer-out leg on the account it leaves, a transfer-in leg on the
// one it reaches. Both legs carry the same dates and status, so they settle
// together; neither counts as an income or an expense of the companies.
//
// Paying a credit card invoice is a transfer to the card, tagged with the
// invoice it pays.

// TransferFields is what a transfer is written from.
type TransferFields struct {
	FromAccount int64
	ToAccount   int64
	FromCompany int64
	// ToCompany is the company the money reaches; FromCompany when 0.
	ToCompany      int64
	Amount         int64
	Description    string
	DocumentNumber string
	Notes          string
	Date           int64
	DueDate        int64
	SettledDate    int64
	Status         string
	// CardInvoice is the invoice (YYYYMM) a transfer to a credit card pays;
	// 0 picks the oldest one closed and unpaid by the date.
	CardInvoice int64
}

// TransferFieldsOf is what the transfer of legs out and in was written from.
func TransferFieldsOf(sandbox *api.Sandbox, out erp_db.TransactionRecord, in erp_db.TransactionRecord) TransferFields {
	return TransferFields{
		FromAccount: out.Account, ToAccount: in.Account, FromCompany: out.Company, ToCompany: in.Company,
		Amount: in.Amount, Description: out.Description, DocumentNumber: out.DocumentNumber, Notes: out.Notes,
		Date: out.Date, DueDate: out.DueDate, SettledDate: out.SettledDate, Status: out.Status, CardInvoice: in.CardInvoice,
	}
}

// FindTransfer is the out leg and the in leg of the transfer the
// transaction with id id is a leg of.
func FindTransfer(sandbox *api.Sandbox, id int64) (erp_db.TransactionRecord, erp_db.TransactionRecord, bool) {
	none := erp_db.TransactionRecord{}
	tx, ok := FindTransaction(sandbox, id)
	if !ok || !erpledger.IsTransfer(sandbox, tx.Type) {
		return none, none, false
	}
	other, ok := FindTransaction(sandbox, tx.Counterpart)
	if !ok {
		return none, none, false
	}
	if tx.Type == erpledger.TypeTransferOut {
		return tx, other, true
	}
	return other, tx, true
}

// validateTransfer cleans fields and tells what is wrong with them, for the
// transfer whose legs were out and in before (nil for a new one), answering
// the accounts it leaves and reaches.
func validateTransfer(sandbox *api.Sandbox, fields *TransferFields, out *erp_db.TransactionRecord, in *erp_db.TransactionRecord) (erp_db.AccountRecord, erp_db.AccountRecord, Problem, error) {
	none := erp_db.AccountRecord{}
	fields.Description = clean(sandbox, fields.Description, 200)
	fields.DocumentNumber = clean(sandbox, fields.DocumentNumber, 60)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	var fromAccount, toAccount, fromCompany, toCompany int64
	if out != nil && in != nil {
		fromAccount, toAccount, fromCompany, toCompany = out.Account, in.Account, out.Company, in.Company
	}
	if fields.ToCompany == 0 {
		fields.ToCompany = fields.FromCompany
	}
	if problem := checkCompany(sandbox, "from-company", fields.FromCompany, fromCompany); Failed(sandbox, problem) {
		return none, none, problem, nil
	}
	if problem := checkCompany(sandbox, "to-company", fields.ToCompany, toCompany); Failed(sandbox, problem) {
		return none, none, problem, nil
	}
	from, problem := checkAccount(sandbox, "from-account", fields.FromAccount, fromAccount)
	if Failed(sandbox, problem) {
		return none, none, problem, nil
	}
	to, problem := checkAccount(sandbox, "to-account", fields.ToAccount, toAccount)
	if Failed(sandbox, problem) {
		return none, none, problem, nil
	}
	if fields.FromAccount == fields.ToAccount && fields.FromCompany == fields.ToCompany {
		return none, none, invalid(sandbox, "to-account", "A transfer moves money to another account, or to another company."), nil
	}
	if fields.Amount <= 0 {
		return none, none, invalid(sandbox, "amount", "Enter an amount greater than zero."), nil
	}
	if fields.Amount > MaxAmount {
		return none, none, invalid(sandbox, "amount", "The amount is too large."), nil
	}
	if fields.Description == "" {
		fields.Description = "Transfer"
		if erpledger.IsCard(sandbox, to) {
			fields.Description = "Credit card payment"
		}
	}
	if erpledger.IsCard(sandbox, from) {
		fields.Status = erpledger.StatusSettled
		fields.SettledDate = fields.Date
	}
	if problem := checkDates(sandbox, fields.Date, &fields.DueDate, &fields.Status, &fields.SettledDate); Failed(sandbox, problem) {
		return none, none, problem, nil
	}
	if erpledger.IsCard(sandbox, to) {
		if fields.CardInvoice == 0 {
			txs, err := Transactions(sandbox)
			if err != nil {
				return none, none, Problem{}, err
			}
			fields.CardInvoice = erpledger.OldestUnpaidInvoice(sandbox, to, txs, fields.DueDate)
			if fields.CardInvoice == 0 {
				fields.CardInvoice = erpledger.InvoiceFor(sandbox, to, fields.DueDate)
			}
		}
		if !erpdate.Valid(sandbox, fields.CardInvoice*100+1) {
			return none, none, invalid(sandbox, "card-invoice", "Enter the invoice as a valid month."), nil
		}
	} else {
		fields.CardInvoice = 0
	}
	return from, to, Problem{}, nil
}

// transferLegs are the inputs of the out leg and the in leg of the transfer
// fields describes, from the account from to the account to, written by the
// user with id userId.
func transferLegs(sandbox *api.Sandbox, fields TransferFields, from erp_db.AccountRecord, to erp_db.AccountRecord, userId int64) (erp_db.TransactionInput, erp_db.TransactionInput) {
	now := nowSeconds(sandbox)
	out := erp_db.TransactionInput{
		Company: fields.FromCompany, Account: fields.FromAccount, Type: erpledger.TypeTransferOut, Amount: -fields.Amount,
		Description: fields.Description, DocumentNumber: fields.DocumentNumber, Notes: fields.Notes,
		Date: fields.Date, DueDate: fields.DueDate, SettledDate: fields.SettledDate, Status: fields.Status,
		CardInvoice: fields.CardInvoice, CreatedAt: now, UpdatedAt: now, CreatedBy: userId,
	}
	in := out
	in.Company, in.Account, in.Type, in.Amount = fields.ToCompany, fields.ToAccount, erpledger.TypeTransferIn, fields.Amount
	if erpledger.IsCard(sandbox, from) {
		out.CardInvoice = erpledger.InvoiceFor(sandbox, from, fields.Date)
	}
	return out, in
}

// AddTransfer writes a new transfer from fields as the user with id userId,
// answering its out leg and its in leg.
func AddTransfer(sandbox *api.Sandbox, fields TransferFields, userId int64) (erp_db.TransactionRecord, erp_db.TransactionRecord, Problem, error) {
	defer lock(sandbox)()
	return addTransfer(sandbox, fields, userId)
}

// addTransfer is AddTransfer under a lock already taken.
func addTransfer(sandbox *api.Sandbox, fields TransferFields, userId int64) (erp_db.TransactionRecord, erp_db.TransactionRecord, Problem, error) {
	none := erp_db.TransactionRecord{}
	from, to, problem, err := validateTransfer(sandbox, &fields, nil, nil)
	if err != nil || Failed(sandbox, problem) {
		return none, none, problem, err
	}
	defer Invalidate(sandbox)
	outInput, inInput := transferLegs(sandbox, fields, from, to, userId)
	out, err := db(sandbox).AddTransaction(outInput)
	if err != nil {
		return none, none, Problem{}, err
	}
	inInput.Counterpart = out.Id
	in, err := db(sandbox).AddTransaction(inInput)
	if err != nil {
		removeAll(sandbox, []erp_db.TransactionRecord{out})
		return none, none, Problem{}, err
	}
	if err := db(sandbox).SetTransactionCounterpart(out.Id, in.Id); err != nil {
		removeAll(sandbox, []erp_db.TransactionRecord{out, in})
		return none, none, Problem{}, err
	}
	out.Counterpart = in.Id
	return out, in, Problem{}, nil
}

// SetTransfer rewrites both legs of the transfer the transaction with id id
// is a leg of, from fields.
func SetTransfer(sandbox *api.Sandbox, id int64, fields TransferFields) (erp_db.TransactionRecord, erp_db.TransactionRecord, Problem, error) {
	defer lock(sandbox)()
	none := erp_db.TransactionRecord{}
	out, in, ok := FindTransfer(sandbox, id)
	if !ok {
		return none, none, notFound(sandbox, "That transfer no longer exists."), nil
	}
	from, to, problem, err := validateTransfer(sandbox, &fields, &out, &in)
	if err != nil || Failed(sandbox, problem) {
		return none, none, problem, err
	}
	defer Invalidate(sandbox)
	outInput, inInput := transferLegs(sandbox, fields, from, to, 0)
	for _, leg := range []struct {
		id    int64
		input erp_db.TransactionInput
	}{{out.Id, outInput}, {in.Id, inInput}} {
		if err := writeLeg(sandbox, leg.id, leg.input); err != nil {
			return none, none, Problem{}, err
		}
	}
	out, in, _ = FindTransfer(sandbox, out.Id)
	return out, in, Problem{}, nil
}

// writeLeg rewrites the transfer leg with id id from input.
func writeLeg(sandbox *api.Sandbox, id int64, input erp_db.TransactionInput) error {
	store := db(sandbox)
	for _, write := range []error{
		store.SetTransactionCompany(id, input.Company),
		store.SetTransactionAccount(id, input.Account),
		store.SetTransactionAmount(id, input.Amount),
		store.SetTransactionDescription(id, input.Description),
		store.SetTransactionDocumentNumber(id, input.DocumentNumber),
		store.SetTransactionNotes(id, input.Notes),
		store.SetTransactionDate(id, input.Date),
		store.SetTransactionDueDate(id, input.DueDate),
		store.SetTransactionSettledDate(id, input.SettledDate),
		store.SetTransactionStatus(id, input.Status),
		store.SetTransactionCardInvoice(id, input.CardInvoice),
		store.SetTransactionUpdatedAt(id, nowSeconds(sandbox)),
	} {
		if write != nil {
			return write
		}
	}
	return nil
}

// InvoicePayment is what paying one invoice of a credit card wrote: one
// transfer per company with something left to pay on it.
type InvoicePayment struct {
	Transfers []erp_db.TransactionRecord
	// Paid is the sum of the transfers.
	Paid int64
}

// PayInvoice pays what is left of the invoice of month (YYYYMM) of the card
// with id card from the account with id from — the card's payment account
// when 0 — on date, settled or pending as status says. Each company pays its
// own share of the invoice, so the card's balance of every company returns to
// zero: one transfer per company that has something left to pay on it.
func PayInvoice(sandbox *api.Sandbox, card int64, month int64, from int64, date int64, status string, userId int64) (InvoicePayment, Problem, error) {
	defer lock(sandbox)()
	account, ok := FindAccount(sandbox, card)
	if !ok || !erpledger.IsCard(sandbox, account) {
		return InvoicePayment{}, notFound(sandbox, "That credit card no longer exists."), nil
	}
	if !erpdate.Valid(sandbox, month*100+1) {
		return InvoicePayment{}, invalid(sandbox, "month", "Enter the invoice as a valid month."), nil
	}
	if from == 0 {
		from = account.PaymentAccount
	}
	if from == 0 {
		return InvoicePayment{}, invalid(sandbox, "from-account", "Choose the account that pays the invoice."), nil
	}
	if date == 0 {
		date = erpdate.Today(sandbox)
	}
	if status == "" {
		status = erpledger.StatusSettled
	}
	txs, err := Transactions(sandbox)
	if err != nil {
		return InvoicePayment{}, Problem{}, err
	}
	shares := erpledger.InvoiceShares(sandbox, account, txs, month)
	payment := InvoicePayment{}
	for _, share := range shares {
		if share.Remaining <= 0 {
			continue
		}
		fields := TransferFields{
			FromAccount: from, ToAccount: card, FromCompany: share.Company, ToCompany: share.Company, Amount: share.Remaining,
			Description: "Credit card payment", Date: date, DueDate: date, Status: status, CardInvoice: month,
		}
		_, in, problem, err := addTransfer(sandbox, fields, userId)
		if err != nil || Failed(sandbox, problem) {
			return payment, problem, err
		}
		payment.Transfers = append(payment.Transfers, in)
		payment.Paid += in.Amount
	}
	if len(payment.Transfers) == 0 {
		return payment, conflict(sandbox, "month", "Nothing is left to pay on this invoice."), nil
	}
	return payment, Problem{}, nil
}
