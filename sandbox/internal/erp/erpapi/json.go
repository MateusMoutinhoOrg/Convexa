package erpapi

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// The documents below are what the ERP api answers: every key in
// kebab-case and in English whatever the language, every amount in signed
// cents, every date spelled 2026-10-12 — "" for none — and every month
// 2026-10. A record names the records it points at by id; 0 is none.
// AddItemToObject copies a child as it is when added, so every child is
// built whole before it is added to its parent.

// Object is a new, empty JSON object.
func Object(sandbox *api.Sandbox) *serializabledeps.SerializableObject {
	return sandbox.Deps.SerializableDeps.CreateObject()
}

// Array is a new, empty JSON array.
func Array(sandbox *api.Sandbox) *serializabledeps.SerializableObject {
	return sandbox.Deps.SerializableDeps.CreateArray()
}

// Write answers document under status.
func Write(sandbox *api.Sandbox, response *serverdeps.Response, status int, document *serializabledeps.SerializableObject) error {
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, status, document)
}

// One answers {key: item} under status.
func One(sandbox *api.Sandbox, response *serverdeps.Response, status int, key string, item *serializabledeps.SerializableObject) error {
	document := Object(sandbox)
	document.AddItemToObject(key, item)
	return Write(sandbox, response, status, document)
}

// Ok answers {"status": "ok"}, what a removal has nothing else to say with.
func Ok(sandbox *api.Sandbox, response *serverdeps.Response) error {
	document := Object(sandbox)
	document.AddItemToObject("status", "ok")
	return Write(sandbox, response, api.StatusOK, document)
}

// Refuse answers problem — the field at fault and why, in the language the
// call asked for — under its status.
func Refuse(sandbox *api.Sandbox, props *routeprops.RouteProps, problem erpstore.Problem) error {
	status := problem.Status
	if status == 0 {
		status = api.StatusBadRequest
	}
	return Fail(sandbox, props, status, problem.Field, problem.Message)
}

// Date reads an ISO date the api was sent under field, 0 for "", refusing
// one that spells no date.
func Date(sandbox *api.Sandbox, props *routeprops.RouteProps, field string, text string) (int64, error) {
	date, ok := erpdate.ParseOptionalISO(sandbox, text)
	if !ok {
		return 0, Fail(sandbox, props, api.StatusBadRequest, field, "Enter a valid date, spelled 2026-10-12.")
	}
	return date, nil
}

// Month reads a month the api was sent under field, spelled 2026-10, 0 for
// "", refusing one that spells no month.
func Month(sandbox *api.Sandbox, props *routeprops.RouteProps, field string, text string) (int64, error) {
	if sandbox.Deps.StringsDeps.TrimSpace(text) == "" {
		return 0, nil
	}
	month, ok := erpdate.ParseMonth(sandbox, text)
	if !ok {
		return 0, Fail(sandbox, props, api.StatusBadRequest, field, "Enter a valid month, spelled 2026-10.")
	}
	return month, nil
}

// CompanyJSON is a company.
func CompanyJSON(sandbox *api.Sandbox, company erp_db.CompanyRecord) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", company.Id)
	object.AddItemToObject("name", company.Name)
	object.AddItemToObject("document", company.Document)
	object.AddItemToObject("notes", company.Notes)
	object.AddItemToObject("status", company.Status)
	object.AddItemToObject("created-at", company.CreatedAt)
	return object
}

// AccountJSON is an account; balance, when not nil, adds its balance today,
// what its pending transactions will move, and the share of each company.
func AccountJSON(sandbox *api.Sandbox, account erp_db.AccountRecord, total int64, pending int64, shares map[int64]int64) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", account.Id)
	object.AddItemToObject("name", account.Name)
	object.AddItemToObject("kind", account.Kind)
	object.AddItemToObject("institution", account.Institution)
	object.AddItemToObject("notes", account.Notes)
	object.AddItemToObject("status", account.Status)
	object.AddItemToObject("credit-limit", account.CreditLimit)
	object.AddItemToObject("closing-day", account.ClosingDay)
	object.AddItemToObject("due-day", account.DueDay)
	object.AddItemToObject("payment-account", account.PaymentAccount)
	object.AddItemToObject("created-at", account.CreatedAt)
	if shares != nil {
		object.AddItemToObject("balance", total)
		object.AddItemToObject("pending", pending)
		companies := Array(sandbox)
		ids := []int64{}
		for company := range shares {
			ids = append(ids, company)
		}
		sandbox.Deps.SortDeps.Slice(ids, func(i int, j int) bool { return ids[i] < ids[j] })
		for _, company := range ids {
			share := Object(sandbox)
			share.AddItemToObject("company", company)
			share.AddItemToObject("balance", shares[company])
			companies.AddItemToArray(share)
		}
		object.AddItemToObject("balances", companies)
	}
	return object
}

// CategoryJSON is a category, path its full name.
func CategoryJSON(sandbox *api.Sandbox, category erp_db.CategoryRecord, path string) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", category.Id)
	object.AddItemToObject("name", category.Name)
	object.AddItemToObject("path", path)
	object.AddItemToObject("type", category.Type)
	object.AddItemToObject("parent", category.Parent)
	object.AddItemToObject("status", category.Status)
	object.AddItemToObject("created-at", category.CreatedAt)
	return object
}

// ThirdPartyJSON is a third party.
func ThirdPartyJSON(sandbox *api.Sandbox, party erp_db.ThirdPartyRecord) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", party.Id)
	object.AddItemToObject("name", party.Name)
	object.AddItemToObject("role", party.Role)
	object.AddItemToObject("document", party.Document)
	object.AddItemToObject("email", party.Email)
	object.AddItemToObject("phone", party.Phone)
	object.AddItemToObject("address", party.Address)
	object.AddItemToObject("notes", party.Notes)
	object.AddItemToObject("status", party.Status)
	object.AddItemToObject("created-at", party.CreatedAt)
	return object
}

// TransactionJSON is a transaction, today telling whether it is overdue.
func TransactionJSON(sandbox *api.Sandbox, tx erp_db.TransactionRecord, today int64) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", tx.Id)
	object.AddItemToObject("company", tx.Company)
	object.AddItemToObject("account", tx.Account)
	object.AddItemToObject("category", tx.Category)
	object.AddItemToObject("third-party", tx.ThirdParty)
	object.AddItemToObject("type", tx.Type)
	object.AddItemToObject("amount", tx.Amount)
	object.AddItemToObject("description", tx.Description)
	object.AddItemToObject("document-number", tx.DocumentNumber)
	object.AddItemToObject("notes", tx.Notes)
	object.AddItemToObject("date", erpdate.ISO(sandbox, tx.Date))
	object.AddItemToObject("due-date", erpdate.ISO(sandbox, tx.DueDate))
	object.AddItemToObject("settled-date", erpdate.ISO(sandbox, tx.SettledDate))
	object.AddItemToObject("status", tx.Status)
	object.AddItemToObject("overdue", tx.Status == erpledger.StatusPending && erpledger.ExpectedDate(sandbox, tx) < today)
	object.AddItemToObject("counterpart", tx.Counterpart)
	object.AddItemToObject("installment-group", tx.InstallmentGroup)
	object.AddItemToObject("installment-number", tx.InstallmentNumber)
	object.AddItemToObject("installment-count", tx.InstallmentCount)
	object.AddItemToObject("recurrence", tx.Recurrence)
	object.AddItemToObject("occurrence", erpdate.ISO(sandbox, tx.Occurrence))
	object.AddItemToObject("card-invoice", erpdate.ISOMonth(sandbox, tx.CardInvoice))
	object.AddItemToObject("created-at", tx.CreatedAt)
	object.AddItemToObject("updated-at", tx.UpdatedAt)
	object.AddItemToObject("created-by", tx.CreatedBy)
	return object
}

// TransactionsJSON is an array of txs.
func TransactionsJSON(sandbox *api.Sandbox, txs []erp_db.TransactionRecord, today int64) *serializabledeps.SerializableObject {
	array := Array(sandbox)
	for _, tx := range txs {
		array.AddItemToArray(TransactionJSON(sandbox, tx, today))
	}
	return array
}

// RecurrenceJSON is a recurrence, next its next occurrence (0 for none).
func RecurrenceJSON(sandbox *api.Sandbox, recurrence erp_db.RecurrenceRecord, next int64) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", recurrence.Id)
	object.AddItemToObject("company", recurrence.Company)
	object.AddItemToObject("account", recurrence.Account)
	object.AddItemToObject("category", recurrence.Category)
	object.AddItemToObject("third-party", recurrence.ThirdParty)
	object.AddItemToObject("type", recurrence.Type)
	object.AddItemToObject("amount", recurrence.Amount)
	object.AddItemToObject("description", recurrence.Description)
	object.AddItemToObject("notes", recurrence.Notes)
	object.AddItemToObject("frequency", recurrence.Frequency)
	object.AddItemToObject("interval", recurrence.Interval)
	object.AddItemToObject("start-date", erpdate.ISO(sandbox, recurrence.StartDate))
	object.AddItemToObject("end-date", erpdate.ISO(sandbox, recurrence.EndDate))
	object.AddItemToObject("generated-until", erpdate.ISO(sandbox, recurrence.GeneratedUntil))
	object.AddItemToObject("next-occurrence", erpdate.ISO(sandbox, next))
	object.AddItemToObject("status", recurrence.Status)
	object.AddItemToObject("created-at", recurrence.CreatedAt)
	return object
}

// BudgetJSON is a budget.
func BudgetJSON(sandbox *api.Sandbox, budget erp_db.BudgetRecord) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("id", budget.Id)
	object.AddItemToObject("company", budget.Company)
	object.AddItemToObject("category", budget.Category)
	object.AddItemToObject("period-start", erpdate.ISO(sandbox, budget.PeriodStart))
	object.AddItemToObject("period-end", erpdate.ISO(sandbox, budget.PeriodEnd))
	object.AddItemToObject("amount", budget.Amount)
	object.AddItemToObject("notes", budget.Notes)
	object.AddItemToObject("created-at", budget.CreatedAt)
	return object
}

// InvoiceJSON is a card invoice.
func InvoiceJSON(sandbox *api.Sandbox, card int64, invoice erpledger.Invoice) *serializabledeps.SerializableObject {
	object := Object(sandbox)
	object.AddItemToObject("card", card)
	object.AddItemToObject("month", erpdate.ISOMonth(sandbox, invoice.Month))
	object.AddItemToObject("closing-date", erpdate.ISO(sandbox, invoice.Closing))
	object.AddItemToObject("due-date", erpdate.ISO(sandbox, invoice.Due))
	object.AddItemToObject("charges", invoice.Charges)
	object.AddItemToObject("paid", invoice.Paid)
	object.AddItemToObject("scheduled", invoice.Scheduled)
	object.AddItemToObject("remaining", invoice.Remaining)
	object.AddItemToObject("status", invoice.Status)
	object.AddItemToObject("count", invoice.Count)
	return object
}
