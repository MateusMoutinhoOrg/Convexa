package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
)

// A budget is how much a company plans to spend — or to receive — in one
// category over a period, a month or any other span. The reports set what
// happened against it, and the projected balance can count what is left of
// it as still to come.

// MaxBudgetMonths is the most monthly budgets one form writes at once.
const MaxBudgetMonths = 60

// BudgetFields is what a budget is written from.
type BudgetFields struct {
	Company     int64
	Category    int64
	PeriodStart int64
	PeriodEnd   int64
	// Amount is in cents, positive.
	Amount int64
	Notes  string
	// Months, above 0 on a new budget, writes that many monthly budgets
	// instead, one per month from the month of PeriodStart, each from the
	// first to the last day of its month.
	Months int64
}

// BudgetFieldsOf is what budget was written from.
func BudgetFieldsOf(sandbox *api.Sandbox, budget erp_db.BudgetRecord) BudgetFields {
	return BudgetFields{Company: budget.Company, Category: budget.Category, PeriodStart: budget.PeriodStart, PeriodEnd: budget.PeriodEnd, Amount: budget.Amount, Notes: budget.Notes}
}

// BudgetQuery narrows a listing of the budgets.
type BudgetQuery struct {
	Company  int64
	Category int64
	// From and To keep the budgets whose period overlaps them; 0 is open.
	From int64
	To   int64
}

// Budgets is every budget query keeps, by period then by id.
func Budgets(sandbox *api.Sandbox, query BudgetQuery) ([]erp_db.BudgetRecord, error) {
	all, err := db(sandbox).ListBudgets(erp_db.BudgetFilter{})
	if err != nil {
		return nil, err
	}
	kept := []erp_db.BudgetRecord{}
	for _, budget := range all {
		if query.Company != 0 && budget.Company != query.Company {
			continue
		}
		if query.Category != 0 && budget.Category != query.Category {
			continue
		}
		if (query.From != 0 && budget.PeriodEnd < query.From) || (query.To != 0 && budget.PeriodStart > query.To) {
			continue
		}
		kept = append(kept, budget)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool {
		if kept[i].PeriodStart != kept[j].PeriodStart {
			return kept[i].PeriodStart < kept[j].PeriodStart
		}
		return kept[i].Id < kept[j].Id
	})
	return kept, nil
}

// FindBudget is the budget with id id.
func FindBudget(sandbox *api.Sandbox, id int64) (erp_db.BudgetRecord, bool) {
	if id <= 0 {
		return erp_db.BudgetRecord{}, false
	}
	return db(sandbox).FindBudgetById(id)
}

// validateBudget cleans fields and tells what is wrong with them, for a
// budget that was current before (nil for a new one).
func validateBudget(sandbox *api.Sandbox, fields *BudgetFields, current *erp_db.BudgetRecord) Problem {
	fields.Notes = clean(sandbox, fields.Notes, 2000)
	var currentCompany int64
	if current != nil {
		currentCompany = current.Company
	}
	if problem := checkCompany(sandbox, "company", fields.Company, currentCompany); Failed(sandbox, problem) {
		return problem
	}
	if _, ok := FindCategory(sandbox, fields.Category); !ok {
		return invalid(sandbox, "category", "Choose the category the budget is for.")
	}
	if fields.Amount <= 0 || fields.Amount > MaxAmount {
		return invalid(sandbox, "amount", "Enter an amount greater than zero.")
	}
	if !erpdate.Valid(sandbox, fields.PeriodStart) {
		return invalid(sandbox, "period-start", "Enter a valid start of the period.")
	}
	if fields.Months < 0 || fields.Months > MaxBudgetMonths {
		return invalid(sandbox, "months", "The months must be between 1 and 60.")
	}
	if fields.Months > 0 {
		return Problem{}
	}
	if !erpdate.Valid(sandbox, fields.PeriodEnd) || fields.PeriodEnd < fields.PeriodStart {
		return invalid(sandbox, "period-end", "The end of the period must be a valid date on or after its start.")
	}
	return Problem{}
}

// AddBudgets writes a new budget from fields — or, with fields.Months, one
// per month — answering every budget written.
func AddBudgets(sandbox *api.Sandbox, fields BudgetFields) ([]erp_db.BudgetRecord, Problem, error) {
	defer lock(sandbox)()
	if problem := validateBudget(sandbox, &fields, nil); Failed(sandbox, problem) {
		return nil, problem, nil
	}
	periods := [][2]int64{{fields.PeriodStart, fields.PeriodEnd}}
	if fields.Months > 0 {
		periods = [][2]int64{}
		first := erpdate.MonthStart(sandbox, fields.PeriodStart)
		for index := int64(0); index < fields.Months; index++ {
			start := erpdate.AddMonths(sandbox, first, index)
			periods = append(periods, [2]int64{start, erpdate.MonthEnd(sandbox, start)})
		}
	}
	written := []erp_db.BudgetRecord{}
	for _, period := range periods {
		budget, err := db(sandbox).AddBudget(erp_db.BudgetInput{
			Company: fields.Company, Category: fields.Category, PeriodStart: period[0], PeriodEnd: period[1],
			Amount: fields.Amount, Notes: fields.Notes, CreatedAt: nowSeconds(sandbox),
		})
		if err != nil {
			for _, done := range written {
				_ = db(sandbox).RemoveBudget(done.Id)
			}
			return nil, Problem{}, err
		}
		written = append(written, budget)
	}
	return written, Problem{}, nil
}

// SetBudget rewrites the budget with id id from fields.
func SetBudget(sandbox *api.Sandbox, id int64, fields BudgetFields) (erp_db.BudgetRecord, Problem, error) {
	defer lock(sandbox)()
	current, ok := FindBudget(sandbox, id)
	if !ok {
		return erp_db.BudgetRecord{}, notFound(sandbox, "That budget no longer exists."), nil
	}
	fields.Months = 0
	if problem := validateBudget(sandbox, &fields, &current); Failed(sandbox, problem) {
		return erp_db.BudgetRecord{}, problem, nil
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetBudgetCompany(id, fields.Company),
		store.SetBudgetCategory(id, fields.Category),
		store.SetBudgetPeriodStart(id, fields.PeriodStart),
		store.SetBudgetPeriodEnd(id, fields.PeriodEnd),
		store.SetBudgetAmount(id, fields.Amount),
		store.SetBudgetNotes(id, fields.Notes),
	} {
		if write != nil {
			return erp_db.BudgetRecord{}, Problem{}, write
		}
	}
	budget, _ := FindBudget(sandbox, id)
	return budget, Problem{}, nil
}

// RemoveBudget deletes the budget with id id.
func RemoveBudget(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindBudget(sandbox, id); !ok {
		return notFound(sandbox, "That budget no longer exists."), nil
	}
	return Problem{}, db(sandbox).RemoveBudget(id)
}
