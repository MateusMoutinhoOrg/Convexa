package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
)

// erpstore is where every financial record of Convexa is written: each write
// is validated here, takes the one write lock, and keeps the records that
// point at each other consistent — the two legs of a transfer, the
// installments of a purchase, the occurrences of a recurrence — so a page, an
// api call and the background job all write the same way.
//
// The store under erp_db has no transactions: a write of several records is
// done under the lock, in an order that leaves nothing pointing at a record
// that does not exist, and undone as far as it can be when a step fails.

// The statuses of a registration: a company, an account, a category, a third
// party. An archived one is kept for the records that name it, and offered
// for no new one.
const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

// The roles of a third party.
const (
	RoleCustomer = "customer"
	RoleSupplier = "supplier"
	RoleBoth     = "both"
)

// The statuses of a recurrence.
const (
	RecurrenceActive = "active"
	RecurrencePaused = "paused"
)

// The frequencies of a recurrence.
const (
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
	FrequencyYearly  = "yearly"
)

// Problem is why a record cannot be written or found, worded for a person in
// English: Field names the input at fault, "" for none, and Status is the
// http status an api call is refused with. The zero Problem is no problem.
type Problem struct {
	Status  int
	Field   string
	Message string
}

// invalid is a Problem with the input field, answered 400.
func invalid(sandbox *api.Sandbox, field string, message string) Problem {
	return Problem{Status: api.StatusBadRequest, Field: field, Message: message}
}

// notFound is the Problem of a record that does not exist, answered 404.
func notFound(sandbox *api.Sandbox, message string) Problem {
	return Problem{Status: api.StatusNotFound, Field: "id", Message: message}
}

// conflict is the Problem of a write the other records forbid, answered 409.
func conflict(sandbox *api.Sandbox, field string, message string) Problem {
	return Problem{Status: api.StatusConflict, Field: field, Message: message}
}

// Failed tells whether problem is one.
func Failed(sandbox *api.Sandbox, problem Problem) bool {
	return problem.Message != ""
}

// writing holds one token while a write runs: every write of the store takes
// it, so two writes never interleave, and the cached listing of the
// transactions is never read half-written.
var writing = make(chan struct{}, 1)

// lock takes the write lock and answers the function that releases it.
func lock(sandbox *api.Sandbox) func() {
	writing <- struct{}{}
	return func() { <-writing }
}

// db is the erp database.
func db(sandbox *api.Sandbox) *erp_db.ErpDb {
	return erp_db.New(sandbox)
}

// nowSeconds is the current time in seconds since the Unix epoch.
func nowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1_000_000_000
}

// clean is text without the spaces around it, cut to max bytes.
func clean(sandbox *api.Sandbox, text string, max int) string {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if len(text) > max {
		text = text[:max]
	}
	return text
}

// sameText tells whether a and b spell the same text, regardless of case and
// of the spaces around them.
func sameText(sandbox *api.Sandbox, a string, b string) bool {
	return sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(a)) == sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(b))
}

// lower is text in lower case, for sorting and searching.
func lower(sandbox *api.Sandbox, text string) string {
	return sandbox.Deps.StringsDeps.ToLower(text)
}

// Matches tells whether text holds search, regardless of case; any text
// matches an empty search.
func Matches(sandbox *api.Sandbox, text string, search string) bool {
	search = sandbox.Deps.StringsDeps.TrimSpace(search)
	if search == "" {
		return true
	}
	return sandbox.Deps.StringsDeps.Contains(lower(sandbox, text), lower(sandbox, search))
}

// oneOf tells whether value is one of options.
func oneOf(sandbox *api.Sandbox, value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

// statusOrActive is status, or active when it is "".
func statusOrActive(sandbox *api.Sandbox, status string) string {
	if status == "" {
		return StatusActive
	}
	return status
}

// cached is the listing of every transaction, read once and kept until a
// write changes one: every report and page computes over the whole ledger,
// and the store reads a record field by field, so reading it again on every
// request would cost as many reads as there are fields in the ledger.
var cached struct {
	valid bool
	list  []erp_db.TransactionRecord
}

// reading holds one token while the cached listing is read or replaced.
var reading = make(chan struct{}, 1)

// Invalidate drops the cached listing of the transactions, so the next read
// lists them again. Every write of the store calls it; so does a backup
// restore, which replaces the whole database under it.
func Invalidate(sandbox *api.Sandbox) {
	reading <- struct{}{}
	cached.valid = false
	cached.list = nil
	<-reading
}

// Transactions is every transaction, in no particular order. The slice is
// the caller's own; the records it holds are shared, and never written to.
func Transactions(sandbox *api.Sandbox) ([]erp_db.TransactionRecord, error) {
	reading <- struct{}{}
	defer func() { <-reading }()
	if !cached.valid {
		list, err := db(sandbox).ListTransactions(erp_db.TransactionFilter{})
		if err != nil {
			return nil, err
		}
		cached.list = list
		cached.valid = true
	}
	copied := make([]erp_db.TransactionRecord, len(cached.list))
	copy(copied, cached.list)
	return copied, nil
}
