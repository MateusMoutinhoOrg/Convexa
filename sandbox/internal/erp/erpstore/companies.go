package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
)

// Everything in Convexa belongs to a company: a person who keeps their own
// finances registers themselves as one ("Personal"). Nothing tells a person
// from a business; a company is a name the records are grouped under.

// CompanyFields is what a company is written from.
type CompanyFields struct {
	Name     string
	Document string
	Notes    string
	Status   string
}

// CompanyFieldsOf is what company was written from.
func CompanyFieldsOf(sandbox *api.Sandbox, company erp_db.CompanyRecord) CompanyFields {
	return CompanyFields{Name: company.Name, Document: company.Document, Notes: company.Notes, Status: company.Status}
}

// CompanyQuery narrows a listing of the companies.
type CompanyQuery struct {
	// Search keeps the companies whose name or document holds it.
	Search string
	// Status keeps the companies of one status, "" for every one.
	Status string
}

// Companies is every company query keeps, by name.
func Companies(sandbox *api.Sandbox, query CompanyQuery) ([]erp_db.CompanyRecord, error) {
	all, err := db(sandbox).ListCompanies(erp_db.CompanyFilter{})
	if err != nil {
		return nil, err
	}
	kept := []erp_db.CompanyRecord{}
	for _, company := range all {
		if query.Status != "" && company.Status != query.Status {
			continue
		}
		if !Matches(sandbox, company.Name+" "+company.Document, query.Search) {
			continue
		}
		kept = append(kept, company)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool { return lower(sandbox, kept[i].Name) < lower(sandbox, kept[j].Name) })
	return kept, nil
}

// FindCompany is the company with id id.
func FindCompany(sandbox *api.Sandbox, id int64) (erp_db.CompanyRecord, bool) {
	if id <= 0 {
		return erp_db.CompanyRecord{}, false
	}
	return db(sandbox).FindCompanyById(id)
}

// validateCompany cleans fields and tells what is wrong with them, for the
// company with id id (0 for a new one).
func validateCompany(sandbox *api.Sandbox, id int64, fields *CompanyFields) (Problem, error) {
	fields.Name = clean(sandbox, fields.Name, 120)
	fields.Document = clean(sandbox, fields.Document, 40)
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	fields.Status = statusOrActive(sandbox, fields.Status)
	if fields.Name == "" {
		return invalid(sandbox, "name", "Enter the company's name."), nil
	}
	if !oneOf(sandbox, fields.Status, StatusActive, StatusArchived) {
		return invalid(sandbox, "status", "The status must be active or archived."), nil
	}
	all, err := db(sandbox).ListCompanies(erp_db.CompanyFilter{})
	if err != nil {
		return Problem{}, err
	}
	for _, other := range all {
		if other.Id != id && sameText(sandbox, other.Name, fields.Name) {
			return invalid(sandbox, "name", "Another company already has that name."), nil
		}
	}
	return Problem{}, nil
}

// AddCompany writes a new company from fields.
func AddCompany(sandbox *api.Sandbox, fields CompanyFields) (erp_db.CompanyRecord, Problem, error) {
	defer lock(sandbox)()
	problem, err := validateCompany(sandbox, 0, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.CompanyRecord{}, problem, err
	}
	company, err := db(sandbox).AddCompany(erp_db.CompanyInput{
		Name: fields.Name, Document: fields.Document, Notes: fields.Notes, Status: fields.Status, CreatedAt: nowSeconds(sandbox),
	})
	return company, Problem{}, err
}

// SetCompany rewrites the company with id id from fields.
func SetCompany(sandbox *api.Sandbox, id int64, fields CompanyFields) (erp_db.CompanyRecord, Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindCompany(sandbox, id); !ok {
		return erp_db.CompanyRecord{}, notFound(sandbox, "That company no longer exists."), nil
	}
	problem, err := validateCompany(sandbox, id, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.CompanyRecord{}, problem, err
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetCompanyName(id, fields.Name),
		store.SetCompanyDocument(id, fields.Document),
		store.SetCompanyNotes(id, fields.Notes),
		store.SetCompanyStatus(id, fields.Status),
	} {
		if write != nil {
			return erp_db.CompanyRecord{}, Problem{}, write
		}
	}
	company, _ := FindCompany(sandbox, id)
	return company, Problem{}, nil
}

// RemoveCompany deletes the company with id id, refused while a transaction,
// a recurrence or a budget names it: archive it instead.
func RemoveCompany(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindCompany(sandbox, id); !ok {
		return notFound(sandbox, "That company no longer exists."), nil
	}
	used, err := companyInUse(sandbox, id)
	if err != nil {
		return Problem{}, err
	}
	if used {
		return conflict(sandbox, "id", "This company has transactions, recurrences or budgets, so it cannot be removed. Archive it instead."), nil
	}
	return Problem{}, db(sandbox).RemoveCompany(id)
}

// companyInUse tells whether a transaction, a recurrence or a budget names
// the company with id id.
func companyInUse(sandbox *api.Sandbox, id int64) (bool, error) {
	txs, err := Transactions(sandbox)
	if err != nil {
		return false, err
	}
	for _, tx := range txs {
		if tx.Company == id {
			return true, nil
		}
	}
	recurrences, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{CompanyMin: id, CompanyMax: id})
	if err != nil || len(recurrences) > 0 {
		return len(recurrences) > 0, err
	}
	budgets, err := db(sandbox).ListBudgets(erp_db.BudgetFilter{CompanyMin: id, CompanyMax: id})
	return len(budgets) > 0, err
}
