package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// A recurrence is an income or an expense that repeats — the rent, a salary,
// a subscription — every so many weeks, months or years from its start date.
// Its occurrences up to a horizon ahead of today (Config.RecurrenceHorizonDays)
// are written as pending transactions, so the payables and receivables list
// them before they are due; the ones beyond it count only in the projections.
// GeneratedUntil is the last day already written.

// RecurrenceFields is what a recurrence is written from.
type RecurrenceFields struct {
	Company    int64
	Account    int64
	Category   int64
	ThirdParty int64
	// Type is income or expense.
	Type string
	// Amount is in cents, positive: the type gives its sign.
	Amount      int64
	Description string
	Notes       string
	// Frequency is weekly, monthly or yearly.
	Frequency string
	// Interval is how many weeks, months or years apart the occurrences are.
	Interval  int64
	StartDate int64
	// EndDate is the last day an occurrence may fall on, 0 for never.
	EndDate int64
	// Status is active or paused.
	Status string
}

// RecurrenceFieldsOf is what recurrence was written from.
func RecurrenceFieldsOf(sandbox *api.Sandbox, recurrence erp_db.RecurrenceRecord) RecurrenceFields {
	return RecurrenceFields{
		Company: recurrence.Company, Account: recurrence.Account, Category: recurrence.Category, ThirdParty: recurrence.ThirdParty,
		Type: recurrence.Type, Amount: recurrence.Amount, Description: recurrence.Description, Notes: recurrence.Notes,
		Frequency: recurrence.Frequency, Interval: recurrence.Interval, StartDate: recurrence.StartDate, EndDate: recurrence.EndDate,
		Status: recurrence.Status,
	}
}

// RecurrenceQuery narrows a listing of the recurrences.
type RecurrenceQuery struct {
	Search  string
	Company int64
	Account int64
	Type    string
	Status  string
}

// Recurrences is every recurrence query keeps, by description.
func Recurrences(sandbox *api.Sandbox, query RecurrenceQuery) ([]erp_db.RecurrenceRecord, error) {
	all, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{})
	if err != nil {
		return nil, err
	}
	kept := []erp_db.RecurrenceRecord{}
	for _, recurrence := range all {
		if query.Company != 0 && recurrence.Company != query.Company {
			continue
		}
		if query.Account != 0 && recurrence.Account != query.Account {
			continue
		}
		if query.Type != "" && recurrence.Type != query.Type {
			continue
		}
		if query.Status != "" && recurrence.Status != query.Status {
			continue
		}
		if !Matches(sandbox, recurrence.Description+" "+recurrence.Notes, query.Search) {
			continue
		}
		kept = append(kept, recurrence)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool {
		return lower(sandbox, kept[i].Description) < lower(sandbox, kept[j].Description)
	})
	return kept, nil
}

// FindRecurrence is the recurrence with id id.
func FindRecurrence(sandbox *api.Sandbox, id int64) (erp_db.RecurrenceRecord, bool) {
	if id <= 0 {
		return erp_db.RecurrenceRecord{}, false
	}
	return db(sandbox).FindRecurrenceById(id)
}

// NextOccurrence is the first day the recurrence falls on after date, 0 when
// it ends before.
func NextOccurrence(sandbox *api.Sandbox, recurrence erp_db.RecurrenceRecord, date int64) int64 {
	dates := erpledger.Occurrences(sandbox, recurrence.StartDate, recurrence.Frequency, recurrence.Interval, recurrence.EndDate, date, erpdate.AddMonths(sandbox, date, 12*recurrence.Interval+12))
	if len(dates) == 0 {
		return 0
	}
	return dates[0]
}

// validateRecurrence cleans fields and tells what is wrong with them, for a
// recurrence that was current before (nil for a new one).
func validateRecurrence(sandbox *api.Sandbox, fields *RecurrenceFields, current *erp_db.RecurrenceRecord) Problem {
	fields.Description = clean(sandbox, fields.Description, 200)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	if fields.Status == "" {
		fields.Status = RecurrenceActive
	}
	if fields.Frequency == "" {
		fields.Frequency = FrequencyMonthly
	}
	if fields.Interval == 0 {
		fields.Interval = 1
	}
	var currentCompany, currentAccount int64
	if current != nil {
		currentCompany, currentAccount = current.Company, current.Account
	}
	if !oneOf(sandbox, fields.Type, erpledger.TypeIncome, erpledger.TypeExpense) {
		return invalid(sandbox, "type", "The type must be income or expense.")
	}
	if problem := checkCompany(sandbox, "company", fields.Company, currentCompany); Failed(sandbox, problem) {
		return problem
	}
	if _, problem := checkAccount(sandbox, "account", fields.Account, currentAccount); Failed(sandbox, problem) {
		return problem
	}
	if problem := checkClassification(sandbox, fields.Type, fields.Category, fields.ThirdParty); Failed(sandbox, problem) {
		return problem
	}
	if fields.Amount <= 0 || fields.Amount > MaxAmount {
		return invalid(sandbox, "amount", "Enter an amount greater than zero.")
	}
	if fields.Description == "" {
		return invalid(sandbox, "description", "Enter a description.")
	}
	if !oneOf(sandbox, fields.Frequency, FrequencyWeekly, FrequencyMonthly, FrequencyYearly) {
		return invalid(sandbox, "frequency", "The frequency must be weekly, monthly or yearly.")
	}
	if fields.Interval < 1 || fields.Interval > 120 {
		return invalid(sandbox, "interval", "The interval must be between 1 and 120.")
	}
	if !erpdate.Valid(sandbox, fields.StartDate) {
		return invalid(sandbox, "start-date", "Enter a valid start date.")
	}
	if fields.EndDate != 0 && (!erpdate.Valid(sandbox, fields.EndDate) || fields.EndDate < fields.StartDate) {
		return invalid(sandbox, "end-date", "The end date must be a valid date on or after the start date.")
	}
	if !oneOf(sandbox, fields.Status, RecurrenceActive, RecurrencePaused) {
		return invalid(sandbox, "status", "The status must be active or paused.")
	}
	return Problem{}
}

// AddRecurrence writes a new recurrence from fields and the pending
// transactions of its occurrences up to the horizon.
func AddRecurrence(sandbox *api.Sandbox, fields RecurrenceFields, userId int64) (erp_db.RecurrenceRecord, Problem, error) {
	defer lock(sandbox)()
	if problem := validateRecurrence(sandbox, &fields, nil); Failed(sandbox, problem) {
		return erp_db.RecurrenceRecord{}, problem, nil
	}
	recurrence, err := db(sandbox).AddRecurrence(erp_db.RecurrenceInput{
		Company: fields.Company, Account: fields.Account, Category: fields.Category, ThirdParty: fields.ThirdParty,
		Type: fields.Type, Amount: fields.Amount, Description: fields.Description, Notes: fields.Notes,
		Frequency: fields.Frequency, Interval: fields.Interval, StartDate: fields.StartDate, EndDate: fields.EndDate,
		GeneratedUntil: firstGenerated(sandbox, fields.StartDate), Status: fields.Status, CreatedAt: nowSeconds(sandbox),
	})
	if err != nil {
		return recurrence, Problem{}, err
	}
	_, err = materialize(sandbox, recurrence, Horizon(sandbox), userId)
	recurrence, _ = FindRecurrence(sandbox, recurrence.Id)
	return recurrence, Problem{}, err
}

// firstGenerated is the GeneratedUntil a new recurrence starting on start
// is written with: the day before it starts, or — for one that started in
// an earlier month — the last day of the month before this one, so a
// recurrence registered late writes the occurrences of this month on, and
// none of the months already closed.
func firstGenerated(sandbox *api.Sandbox, start int64) int64 {
	floor := erpdate.AddDays(sandbox, erpdate.MonthStart(sandbox, erpdate.Today(sandbox)), -1)
	before := erpdate.AddDays(sandbox, start, -1)
	if before > floor {
		return before
	}
	return floor
}

// SetRecurrence rewrites the recurrence with id id from fields. Its pending
// occurrences from today on are written again from the new fields; the ones
// already settled, and the overdue ones, stay as they are.
func SetRecurrence(sandbox *api.Sandbox, id int64, fields RecurrenceFields, userId int64) (erp_db.RecurrenceRecord, Problem, error) {
	defer lock(sandbox)()
	current, ok := FindRecurrence(sandbox, id)
	if !ok {
		return erp_db.RecurrenceRecord{}, notFound(sandbox, "That recurrence no longer exists."), nil
	}
	if problem := validateRecurrence(sandbox, &fields, &current); Failed(sandbox, problem) {
		return erp_db.RecurrenceRecord{}, problem, nil
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetRecurrenceCompany(id, fields.Company),
		store.SetRecurrenceAccount(id, fields.Account),
		store.SetRecurrenceCategory(id, fields.Category),
		store.SetRecurrenceThirdParty(id, fields.ThirdParty),
		store.SetRecurrenceType(id, fields.Type),
		store.SetRecurrenceAmount(id, fields.Amount),
		store.SetRecurrenceDescription(id, fields.Description),
		store.SetRecurrenceNotes(id, fields.Notes),
		store.SetRecurrenceFrequency(id, fields.Frequency),
		store.SetRecurrenceInterval(id, fields.Interval),
		store.SetRecurrenceStartDate(id, fields.StartDate),
		store.SetRecurrenceEndDate(id, fields.EndDate),
		store.SetRecurrenceStatus(id, fields.Status),
	} {
		if write != nil {
			return erp_db.RecurrenceRecord{}, Problem{}, write
		}
	}
	if err := dropUpcoming(sandbox, id, true); err != nil {
		return erp_db.RecurrenceRecord{}, Problem{}, err
	}
	recurrence, _ := FindRecurrence(sandbox, id)
	_, err := materialize(sandbox, recurrence, Horizon(sandbox), userId)
	recurrence, _ = FindRecurrence(sandbox, id)
	return recurrence, Problem{}, err
}

// RemoveRecurrence deletes the recurrence with id id and its pending
// occurrences from today on; the transactions it wrote before stay, no longer
// tied to it.
func RemoveRecurrence(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindRecurrence(sandbox, id); !ok {
		return notFound(sandbox, "That recurrence no longer exists."), nil
	}
	if err := dropUpcoming(sandbox, id, false); err != nil {
		return Problem{}, err
	}
	txs, err := Transactions(sandbox)
	if err != nil {
		return Problem{}, err
	}
	for _, tx := range txs {
		if tx.Recurrence == id {
			if err := db(sandbox).SetTransactionRecurrence(tx.Id, 0); err != nil {
				return Problem{}, err
			}
		}
	}
	Invalidate(sandbox)
	return Problem{}, db(sandbox).RemoveRecurrence(id)
}

// dropUpcoming deletes the pending occurrences of the recurrence with id id
// that fall from today on and, with rewind, moves its GeneratedUntil back to
// yesterday, so they are written again.
func dropUpcoming(sandbox *api.Sandbox, id int64, rewind bool) error {
	today := erpdate.Today(sandbox)
	txs, err := Transactions(sandbox)
	if err != nil {
		return err
	}
	defer Invalidate(sandbox)
	for _, tx := range txs {
		if tx.Recurrence == id && tx.Occurrence >= today && tx.Status == erpledger.StatusPending {
			if err := db(sandbox).RemoveTransaction(tx.Id); err != nil {
				return err
			}
		}
	}
	if !rewind {
		return nil
	}
	recurrence, _ := FindRecurrence(sandbox, id)
	yesterday := erpdate.AddDays(sandbox, today, -1)
	until := erpdate.AddDays(sandbox, recurrence.StartDate, -1)
	if yesterday > until {
		until = yesterday
	}
	if recurrence.GeneratedUntil < until {
		until = recurrence.GeneratedUntil
	}
	return db(sandbox).SetRecurrenceGeneratedUntil(id, until)
}

// Horizon is the last day the occurrences of the recurrences are written
// up to: Config.RecurrenceHorizonDays after today.
func Horizon(sandbox *api.Sandbox) int64 {
	return erpdate.AddDays(sandbox, erpdate.Today(sandbox), sandbox.Config.RecurrenceHorizonDays)
}

// materialize writes the transactions of the occurrences of recurrence after
// its GeneratedUntil and up to until — none for a paused one — and moves its
// GeneratedUntil there, answering how many it wrote. An occurrence already
// written, by a run that stopped halfway, is not written twice.
func materialize(sandbox *api.Sandbox, recurrence erp_db.RecurrenceRecord, until int64, userId int64) (int, error) {
	if recurrence.Status != RecurrenceActive || until <= recurrence.GeneratedUntil {
		return 0, nil
	}
	account, ok := FindAccount(sandbox, recurrence.Account)
	if !ok {
		return 0, nil
	}
	dates := erpledger.Occurrences(sandbox, recurrence.StartDate, recurrence.Frequency, recurrence.Interval, recurrence.EndDate, recurrence.GeneratedUntil, until)
	written := 0
	if len(dates) > 0 {
		txs, err := Transactions(sandbox)
		if err != nil {
			return 0, err
		}
		done := map[int64]bool{}
		for _, tx := range txs {
			if tx.Recurrence == recurrence.Id {
				done[tx.Occurrence] = true
			}
		}
		defer Invalidate(sandbox)
		now := nowSeconds(sandbox)
		for _, date := range dates {
			if done[date] {
				continue
			}
			input := erp_db.TransactionInput{
				Company: recurrence.Company, Account: recurrence.Account, Category: recurrence.Category, ThirdParty: recurrence.ThirdParty,
				Type: recurrence.Type, Amount: signed(sandbox, recurrence.Type, recurrence.Amount), Description: recurrence.Description,
				Notes: recurrence.Notes, Date: date, DueDate: date, Status: erpledger.StatusPending,
				Recurrence: recurrence.Id, Occurrence: date, CreatedAt: now, UpdatedAt: now, CreatedBy: userId,
			}
			if erpledger.IsCard(sandbox, account) {
				input.Status = erpledger.StatusSettled
				input.SettledDate = date
				input.CardInvoice = erpledger.InvoiceFor(sandbox, account, date)
			}
			if _, err := db(sandbox).AddTransaction(input); err != nil {
				return written, err
			}
			written++
		}
	}
	return written, db(sandbox).SetRecurrenceGeneratedUntil(recurrence.Id, until)
}

// GenerateRecurrences writes the occurrences of every active recurrence up
// to the horizon, answering how many transactions it wrote. start-server runs
// it when the server starts and every few hours after; a page and the api run
// it on demand.
func GenerateRecurrences(sandbox *api.Sandbox, userId int64) (int, error) {
	defer lock(sandbox)()
	recurrences, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{})
	if err != nil {
		return 0, err
	}
	until := Horizon(sandbox)
	total := 0
	for _, recurrence := range recurrences {
		written, err := materialize(sandbox, recurrence, until, userId)
		total += written
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
