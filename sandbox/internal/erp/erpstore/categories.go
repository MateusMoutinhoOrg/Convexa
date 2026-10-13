package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// A category classifies income or expense — "Rent", "Sales", "Payroll" — and
// may sit under a parent of the same type, so a report or a budget of the
// parent counts its children too.

// CategoryFields is what a category is written from.
type CategoryFields struct {
	Name string
	// Type is income or expense.
	Type string
	// Parent is the id of the category this one sits under, 0 for none.
	Parent int64
	Status string
}

// CategoryFieldsOf is what category was written from.
func CategoryFieldsOf(sandbox *api.Sandbox, category erp_db.CategoryRecord) CategoryFields {
	return CategoryFields{Name: category.Name, Type: category.Type, Parent: category.Parent, Status: category.Status}
}

// CategoryQuery narrows a listing of the categories.
type CategoryQuery struct {
	Search string
	// Type keeps the categories of one type, "" for both.
	Type string
	// Status keeps the categories of one status, "" for every one.
	Status string
}

// Categories is every category query keeps, by type then by path — a child
// right after its parent.
func Categories(sandbox *api.Sandbox, query CategoryQuery) ([]erp_db.CategoryRecord, error) {
	all, err := db(sandbox).ListCategories(erp_db.CategoryFilter{})
	if err != nil {
		return nil, err
	}
	paths := CategoryPaths(sandbox, all)
	kept := []erp_db.CategoryRecord{}
	for _, category := range all {
		if query.Type != "" && category.Type != query.Type {
			continue
		}
		if query.Status != "" && category.Status != query.Status {
			continue
		}
		if !Matches(sandbox, paths[category.Id], query.Search) {
			continue
		}
		kept = append(kept, category)
	}
	sandbox.Deps.SortDeps.Slice(kept, func(i int, j int) bool {
		if kept[i].Type != kept[j].Type {
			return kept[i].Type == erpledger.TypeIncome
		}
		return lower(sandbox, paths[kept[i].Id]) < lower(sandbox, paths[kept[j].Id])
	})
	return kept, nil
}

// CategoryPaths is the full name of every category of all, keyed by id:
// "Housing › Rent" for Rent under Housing.
func CategoryPaths(sandbox *api.Sandbox, all []erp_db.CategoryRecord) map[int64]string {
	byId := map[int64]erp_db.CategoryRecord{}
	for _, category := range all {
		byId[category.Id] = category
	}
	paths := map[int64]string{}
	for _, category := range all {
		path := category.Name
		seen := map[int64]bool{category.Id: true}
		parent := category.Parent
		for parent != 0 && !seen[parent] {
			above, ok := byId[parent]
			if !ok {
				break
			}
			path = above.Name + " › " + path
			seen[parent] = true
			parent = above.Parent
		}
		paths[category.Id] = path
	}
	return paths
}

// CategorySubtree is the id of the category with id root and of every
// category under it, at any depth, among all.
func CategorySubtree(sandbox *api.Sandbox, all []erp_db.CategoryRecord, root int64) map[int64]bool {
	subtree := map[int64]bool{root: true}
	for grew := true; grew; {
		grew = false
		for _, category := range all {
			if !subtree[category.Id] && category.Parent != 0 && subtree[category.Parent] {
				subtree[category.Id] = true
				grew = true
			}
		}
	}
	return subtree
}

// FindCategory is the category with id id.
func FindCategory(sandbox *api.Sandbox, id int64) (erp_db.CategoryRecord, bool) {
	if id <= 0 {
		return erp_db.CategoryRecord{}, false
	}
	return db(sandbox).FindCategoryById(id)
}

// validateCategory cleans fields and tells what is wrong with them, for the
// category with id id (0 for a new one).
func validateCategory(sandbox *api.Sandbox, id int64, fields *CategoryFields) (Problem, error) {
	fields.Name = clean(sandbox, fields.Name, 120)
	fields.Status = statusOrActive(sandbox, fields.Status)
	if fields.Name == "" {
		return invalid(sandbox, "name", "Enter the category's name."), nil
	}
	if !oneOf(sandbox, fields.Type, erpledger.TypeIncome, erpledger.TypeExpense) {
		return invalid(sandbox, "type", "The type must be income or expense."), nil
	}
	if !oneOf(sandbox, fields.Status, StatusActive, StatusArchived) {
		return invalid(sandbox, "status", "The status must be active or archived."), nil
	}
	all, err := db(sandbox).ListCategories(erp_db.CategoryFilter{})
	if err != nil {
		return Problem{}, err
	}
	if fields.Parent != 0 {
		parent, ok := FindCategory(sandbox, fields.Parent)
		if !ok {
			return invalid(sandbox, "parent", "The parent category does not exist."), nil
		}
		if parent.Type != fields.Type {
			return invalid(sandbox, "parent", "A category sits only under a parent of the same type."), nil
		}
		if id != 0 && CategorySubtree(sandbox, all, id)[fields.Parent] {
			return invalid(sandbox, "parent", "A category cannot sit under itself or under one of its own subcategories."), nil
		}
	}
	for _, other := range all {
		if other.Id != id && other.Parent == fields.Parent && other.Type == fields.Type && sameText(sandbox, other.Name, fields.Name) {
			return invalid(sandbox, "name", "Another category of this type already has that name there."), nil
		}
	}
	if id != 0 {
		current, _ := FindCategory(sandbox, id)
		if current.Type != fields.Type {
			used, err := categoryInUse(sandbox, id, true)
			if err != nil {
				return Problem{}, err
			}
			if used {
				return invalid(sandbox, "type", "A category that is in use, or has subcategories, cannot change its type."), nil
			}
		}
	}
	return Problem{}, nil
}

// AddCategory writes a new category from fields.
func AddCategory(sandbox *api.Sandbox, fields CategoryFields) (erp_db.CategoryRecord, Problem, error) {
	defer lock(sandbox)()
	problem, err := validateCategory(sandbox, 0, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.CategoryRecord{}, problem, err
	}
	category, err := db(sandbox).AddCategory(erp_db.CategoryInput{
		Name: fields.Name, Type: fields.Type, Parent: fields.Parent, Status: fields.Status, CreatedAt: nowSeconds(sandbox),
	})
	return category, Problem{}, err
}

// SetCategory rewrites the category with id id from fields.
func SetCategory(sandbox *api.Sandbox, id int64, fields CategoryFields) (erp_db.CategoryRecord, Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindCategory(sandbox, id); !ok {
		return erp_db.CategoryRecord{}, notFound(sandbox, "That category no longer exists."), nil
	}
	problem, err := validateCategory(sandbox, id, &fields)
	if err != nil || Failed(sandbox, problem) {
		return erp_db.CategoryRecord{}, problem, err
	}
	store := db(sandbox)
	for _, write := range []error{
		store.SetCategoryName(id, fields.Name),
		store.SetCategoryType(id, fields.Type),
		store.SetCategoryParent(id, fields.Parent),
		store.SetCategoryStatus(id, fields.Status),
	} {
		if write != nil {
			return erp_db.CategoryRecord{}, Problem{}, write
		}
	}
	category, _ := FindCategory(sandbox, id)
	return category, Problem{}, nil
}

// RemoveCategory deletes the category with id id, refused while it has
// subcategories or a transaction, a recurrence or a budget names it.
func RemoveCategory(sandbox *api.Sandbox, id int64) (Problem, error) {
	defer lock(sandbox)()
	if _, ok := FindCategory(sandbox, id); !ok {
		return notFound(sandbox, "That category no longer exists."), nil
	}
	used, err := categoryInUse(sandbox, id, true)
	if err != nil {
		return Problem{}, err
	}
	if used {
		return conflict(sandbox, "id", "This category has subcategories, transactions, recurrences or budgets, so it cannot be removed. Archive it instead."), nil
	}
	return Problem{}, db(sandbox).RemoveCategory(id)
}

// categoryInUse tells whether a transaction, a recurrence or a budget names
// the category with id id — or, with children, a subcategory sits under it.
func categoryInUse(sandbox *api.Sandbox, id int64, children bool) (bool, error) {
	if children {
		under, err := db(sandbox).ListCategories(erp_db.CategoryFilter{ParentMin: id, ParentMax: id})
		if err != nil || len(under) > 0 {
			return len(under) > 0, err
		}
	}
	txs, err := Transactions(sandbox)
	if err != nil {
		return false, err
	}
	for _, tx := range txs {
		if tx.Category == id {
			return true, nil
		}
	}
	recurrences, err := db(sandbox).ListRecurrences(erp_db.RecurrenceFilter{CategoryMin: id, CategoryMax: id})
	if err != nil || len(recurrences) > 0 {
		return len(recurrences) > 0, err
	}
	budgets, err := db(sandbox).ListBudgets(erp_db.BudgetFilter{CategoryMin: id, CategoryMax: id})
	return len(budgets) > 0, err
}
