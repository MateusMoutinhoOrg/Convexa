package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
)

// A third party is whoever a company pays or is paid by: a customer, a
// supplier, or both at once.

// ThirdPartyFields is what a third party is written from.
type ThirdPartyFields struct {
	Name string
	// Role is customer, supplier or both.
	Role     string
	Document string
	Email    string
	Phone    string
	Address  string
	Notes    string
	Status   string
}

// ThirdPartyFieldsOf is what party was written from.
func ThirdPartyFieldsOf(sandbox *api.Sandbox, party erp_db.ThirdPartyRecord) ThirdPartyFields {
	return ThirdPartyFields{Name: party.Name, Role: party.Role, Document: party.Document, Email: party.Email, Phone: party.Phone, Address: party.Address, Notes: party.Notes, Status: party.Status}
}

// ThirdPartyQuery narrows a listing of the third parties.
type ThirdPartyQuery struct {
	// Search keeps the parties whose name, document, email or phone holds it.
	Search string
	// Role keeps the customers (customer), the suppliers (supplier) — both
	// counting the parties that are both — or "" every one.
	Role string
	// Status keeps the parties of one status, "" for every one.
	Status string
}

// ThirdParties is every third party query keeps, by name.
func ThirdParties(sandbox *api.Sandbox, query ThirdPartyQuery) ([]erp_db.ThirdPartyRecord, error) {
	all, err := db(sandbox).ListThirdParties(erp_db.ThirdPartyFilter{})
	if err != nil {
		return nil, err
	}
	kept := []erp_db.ThirdPartyRecord{}
	for _, party := range all {
		if query.Role != "" && party.Role != query.Role && party.Role != RoleBoth {
			continue
		}
		if query.Status != "" && party.Status != query.Status {
			continue
		}
		if !Matches(sandbox, party.Name+" "+party.Document+" "+party.Email+" "+party.Phone, query.Search) {
			continue
		}
		kept = append(kept, party)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool { return lower(sandbox, kept[i].Name) < lower(sandbox, kept[j].Name) })
	return kept, nil
}

// FindThirdParty is the third party with id id.
func FindThirdParty(sandbox *api.Sandbox, id int64) (erp_db.ThirdPartyRecord, bool) {
	if id <= 0 {
		return erp_db.ThirdPartyRecord{}, false
	}
	return db(sandbox).FindThirdPartyById(id)
}

// validateThirdParty cleans fields and tells what is wrong with them.
func validateThirdParty(sandbox *api.Sandbox, fields *ThirdPartyFields) Problem {
	fields.Name = clean(sandbox, fields.Name, 160)
	fields.Document = clean(sandbox, fields.Document, 40)
	fields.Email = clean(sandbox, fields.Email, 254)
	fields.Phone = clean(sandbox, fields.Phone, 40)
	fields.Address = clean(sandbox, fields.Address, 500)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	fields.Status = statusOrActive(sandbox, fields.Status)
	if fields.Role == "" {
		fields.Role = RoleBoth
	}
	if fields.Name == "" {
		return invalid(sandbox, "name", "Enter the third party's name.")
	}
	if !oneOf(sandbox, fields.Role, RoleCustomer, RoleSupplier, RoleBoth) {
		return invalid(sandbox, "role", "The role must be customer, supplier or both.")
	}
	if !oneOf(sandbox, fields.Status, StatusActive, StatusArchived) {
		return invalid(sandbox, "status", "The status must be active or archived.")
	}
	if fields.Email != "" {
		at := sandbox.Deps.StringsDeps.LastIndex(fields.Email, "@")
		if at < 1 || at == len(fields.Email)-1 || sandbox.Deps.StringsDeps.ContainsAny(fields.Email, " \t") {
			return invalid(sandbox, "email", "Enter a valid email, or leave it blank.")
		}
	}
	return Problem{}
}

// AddThirdParty writes a new third party from fields.
func AddThirdParty(sandbox *api.Sandbox, fields ThirdPartyFields) (erp_db.ThirdPartyRecord, Problem, error) {
	defer lock(sandbox)()
	if problem := validateThirdParty(sandbox, &fields); Failed(sandbox, problem) {
		return erp_db.ThirdPartyRecord{}, problem, nil
	}
	party, err := db(sandbox).AddThirdParty(erp_db.ThirdPartyInput{
		Name: fields.Name, Role: fields.Role, Document: fields.Document, Email: fields.Email, Phone: fields.Phone,
		Address: fields.Address, Notes: fields.Notes, Status: fields.Status, CreatedAt: nowSeconds(sandbox),
	})
	return party, Problem{}, err
}

// SetThirdParty rewrites the third party with id id from fields.
func SetThirdParty(sandbox *api.Sandbox, id int64, fields ThirdPartyFields) (erp_db.ThirdPartyRecord, Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindThirdParty(sandbox, id); !ok {
		return erp_db.ThirdPartyRecord{}, notFound(sandbox, "That third party no longer exists."), nil
	}
	if problem := validateThirdParty(sandbox, &fields); Failed(sandbox, problem) {
		return erp_db.ThirdPartyRecord{}, problem, nil
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetThirdPartyName(id, fields.Name),
		store.SetThirdPartyRole(id, fields.Role),
		store.SetThirdPartyDocument(id, fields.Document),
		store.SetThirdPartyEmail(id, fields.Email),
		store.SetThirdPartyPhone(id, fields.Phone),
		store.SetThirdPartyAddress(id, fields.Address),
		store.SetThirdPartyNotes(id, fields.Notes),
		store.SetThirdPartyStatus(id, fields.Status),
	} {
		if write != nil {
			return erp_db.ThirdPartyRecord{}, Problem{}, write
		}
	}
	party, _ := FindThirdParty(sandbox, id)
	return party, Problem{}, nil
}

// RemoveThirdParty deletes the third party with id id, refused while a
// transaction or a recurrence names it.
func RemoveThirdParty(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindThirdParty(sandbox, id); !ok {
		return notFound(sandbox, "That third party no longer exists."), nil
	}
	txs, err := Transactions(sandbox)
	if err != nil {
		return Problem{}, err
	}
	for _, tx := range txs {
		if tx.ThirdParty == id {
			return conflict(sandbox, "id", "This third party has transactions, so it cannot be removed. Archive it instead."), nil
		}
	}
	recurrences, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{ThirdPartyMin: id, ThirdPartyMax: id})
	if err != nil {
		return Problem{}, err
	}
	if len(recurrences) > 0 {
		return conflict(sandbox, "id", "A recurrence uses this third party, so it cannot be removed. Archive it instead."), nil
	}
	return Problem{}, db(sandbox).RemoveThirdParty(id)
}
