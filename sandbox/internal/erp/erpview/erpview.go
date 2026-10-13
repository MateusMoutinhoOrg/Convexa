package erpview

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpmoney"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
)

// erpview holds what the ERP pages share: the options of their selects, the
// sentences of their notices, and the reading of the strings a form posts.
// The words it holds are English; the page translates them.

// Option is one option of a select.
type Option struct {
	Value    string
	Label    string
	Selected bool
	// Group heads the options of an optgroup, when a page groups them.
	Group string
}

// Choice is an option of value labeled label, selected when value is
// selected.
func Choice(sandbox *api.Sandbox, value string, label string, selected string) Option {
	return Option{Value: value, Label: label, Selected: value == selected}
}

// Choices are an option per value, labeled with the value itself — the page
// translates it — and the one of selected selected.
func Choices(sandbox *api.Sandbox, values []string, selected string) []Option {
	options := []Option{}
	for _, value := range values {
		options = append(options, Choice(sandbox, value, value, selected))
	}
	return options
}

// id spells id as an option value, "" for 0.
func id(sandbox *api.Sandbox, value int64) string {
	if value == 0 {
		return ""
	}
	return sandbox.Deps.StringsDeps.FormatInt(value, 10)
}

// Id spells an id the way a form posts it, "" for 0.
func Id(sandbox *api.Sandbox, value int64) string {
	return id(sandbox, value)
}

// CompanyOptions are an option per company of companies — the archived ones
// only when selected — the one with id selected selected.
func CompanyOptions(sandbox *api.Sandbox, companies []erp_db.CompanyRecord, selected int64) []Option {
	options := []Option{}
	for _, company := range companies {
		if company.Status != erpstore.StatusActive && company.Id != selected {
			continue
		}
		options = append(options, Option{Value: id(sandbox, company.Id), Label: company.Name, Selected: company.Id == selected})
	}
	return options
}

// AccountOptions are an option per account of accounts — the archived ones
// only when selected, the cards only with cards — the one with id selected
// selected.
func AccountOptions(sandbox *api.Sandbox, accounts []erp_db.AccountRecord, selected int64, cards bool) []Option {
	options := []Option{}
	for _, account := range accounts {
		if account.Status != erpstore.StatusActive && account.Id != selected {
			continue
		}
		if !cards && erpledger.IsCard(sandbox, account) && account.Id != selected {
			continue
		}
		options = append(options, Option{Value: id(sandbox, account.Id), Label: account.Name, Selected: account.Id == selected})
	}
	return options
}

// CategoryOptions are an option per category of categories of type kind ("" for
// both), by full name, the archived ones only when selected, the one with id
// selected selected; Group says income or expense.
func CategoryOptions(sandbox *api.Sandbox, categories []erp_db.CategoryRecord, kind string, selected int64) []Option {
	paths := erpstore.CategoryPaths(sandbox, categories)
	options := []Option{}
	for _, category := range categories {
		if kind != "" && category.Type != kind {
			continue
		}
		if category.Status != erpstore.StatusActive && category.Id != selected {
			continue
		}
		options = append(options, Option{Value: id(sandbox, category.Id), Label: paths[category.Id], Selected: category.Id == selected, Group: category.Type})
	}
	return options
}

// ThirdPartyOptions are an option per third party of parties, the archived
// ones only when selected, the one with id selected selected.
func ThirdPartyOptions(sandbox *api.Sandbox, parties []erp_db.ThirdPartyRecord, selected int64) []Option {
	options := []Option{}
	for _, party := range parties {
		if party.Status != erpstore.StatusActive && party.Id != selected {
			continue
		}
		options = append(options, Option{Value: id(sandbox, party.Id), Label: party.Name, Selected: party.Id == selected})
	}
	return options
}

// ParseId is the id text spells, 0 for "" or for text that spells none.
func ParseId(sandbox *api.Sandbox, text string) int64 {
	value, err := sandbox.Deps.StringsDeps.ParseInt(sandbox.Deps.StringsDeps.TrimSpace(text), 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

// ParseInt is the integer text spells, 0 for "" or for text that spells none.
func ParseInt(sandbox *api.Sandbox, text string) int64 {
	value, err := sandbox.Deps.StringsDeps.ParseInt(sandbox.Deps.StringsDeps.TrimSpace(text), 10, 64)
	if err != nil {
		return 0
	}
	return value
}

// Field is one value a form posted, and whether it could be read.
type Field struct {
	Name string
	Ok   bool
}

// Reader reads the strings a form posts into the values of a record,
// remembering the first it could not read, so the form is answered again with
// that field named.
type Reader struct {
	// Field and Message say what could not be read; Message is "" when
	// everything was.
	Field   string
	Message string
}

// fail remembers the first field that could not be read.
func (reader *Reader) fail(field string, message string) {
	if reader.Message == "" {
		reader.Field, reader.Message = field, message
	}
}

// Money reads an amount from text, 0 for "".
func Money(sandbox *api.Sandbox, reader *Reader, field string, text string) int64 {
	if sandbox.Deps.StringsDeps.TrimSpace(text) == "" {
		return 0
	}
	cents, ok := erpmoney.Parse(sandbox, text)
	if !ok {
		reader.fail(field, "Enter a valid amount, such as 1234.56.")
	}
	return cents
}

// Date reads an ISO date from text, 0 for "".
func Date(sandbox *api.Sandbox, reader *Reader, field string, text string) int64 {
	date, ok := erpdate.ParseOptionalISO(sandbox, text)
	if !ok {
		reader.fail(field, "Enter a valid date.")
	}
	return date
}

// Month reads a month, spelled 2026-10, from text, 0 for "".
func Month(sandbox *api.Sandbox, reader *Reader, field string, text string) int64 {
	if sandbox.Deps.StringsDeps.TrimSpace(text) == "" {
		return 0
	}
	month, ok := erpdate.ParseMonth(sandbox, text)
	if !ok {
		reader.fail(field, "Enter a valid month.")
	}
	return month
}

// Int reads a whole number from text, 0 for "".
func Int(sandbox *api.Sandbox, reader *Reader, field string, text string) int64 {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if text == "" {
		return 0
	}
	value, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
	if err != nil {
		reader.fail(field, "Enter a whole number.")
	}
	return value
}

// The notice codes a write redirects to its list with: ?notice=added.
const (
	NoticeAdded        = "added"
	NoticeUpdated      = "updated"
	NoticeRemoved      = "removed"
	NoticeNotFound     = "not-found"
	NoticeInUse        = "in-use"
	NoticeSettled      = "settled"
	NoticeUnsettled    = "unsettled"
	NoticePaid         = "paid"
	NoticeNothingToPay = "nothing-to-pay"
	NoticeGenerated    = "generated"
	NoticeRefused      = "refused"
)

// NoticeOf is the notice the code of a redirect stands for.
func NoticeOf(sandbox *api.Sandbox, code string) erppage.Notice {
	switch code {
	case NoticeAdded:
		return erppage.Ok(sandbox, "Saved.")
	case NoticeUpdated:
		return erppage.Ok(sandbox, "Changes saved.")
	case NoticeRemoved:
		return erppage.Ok(sandbox, "Removed.")
	case NoticeSettled:
		return erppage.Ok(sandbox, "Marked as settled.")
	case NoticeUnsettled:
		return erppage.Ok(sandbox, "Marked as pending again.")
	case NoticePaid:
		return erppage.Ok(sandbox, "Invoice paid.")
	case NoticeGenerated:
		return erppage.Ok(sandbox, "The recurring entries are up to date.")
	case NoticeNotFound:
		return erppage.Error(sandbox, "That record no longer exists.")
	case NoticeInUse:
		return erppage.Error(sandbox, "It is in use, so it cannot be removed. Archive it instead.")
	case NoticeNothingToPay:
		return erppage.Error(sandbox, "Nothing is left to pay on this invoice.")
	case NoticeRefused:
		return erppage.Error(sandbox, "That change is not allowed.")
	}
	return erppage.Notice{}
}

// NoticeOfProblem is the notice code a refused remove or action redirects
// with.
func NoticeOfProblem(sandbox *api.Sandbox, problem erpstore.Problem) string {
	switch problem.Status {
	case api.StatusNotFound:
		return NoticeNotFound
	case api.StatusConflict:
		return NoticeInUse
	}
	return NoticeRefused
}

// WithNotice is location with ?notice=code added, or &notice=code when it
// already has a query.
func WithNotice(sandbox *api.Sandbox, location string, code string) string {
	if sandbox.Deps.StringsDeps.Contains(location, "?") {
		return location + "&notice=" + code
	}
	return location + "?notice=" + code
}

// SafeNext is next when it is a path of the ERP pages, fallback otherwise:
// a form redirects back to where it was posted from, never to another site.
func SafeNext(sandbox *api.Sandbox, next string, fallback string) string {
	if !sandbox.Deps.StringsDeps.HasPrefix(next, "/admin/erp/") || sandbox.Deps.StringsDeps.ContainsAny(next, "\\\r\n") {
		return fallback
	}
	return next
}

// Statuses are the statuses of a registration.
func Statuses(sandbox *api.Sandbox) []string {
	return []string{erpstore.StatusActive, erpstore.StatusArchived}
}

// moneyInput is cents as an amount input of lang is filled, "" for 0.
func moneyInput(sandbox *api.Sandbox, lang string, cents int64) string {
	return erpmoney.FormatInput(sandbox, lang, cents)
}

// ProblemOf is the Problem of what reader could not read, the zero Problem
// when it read everything.
func ProblemOf(sandbox *api.Sandbox, reader Reader) erpstore.Problem {
	if reader.Message == "" {
		return erpstore.Problem{}
	}
	return erpstore.Problem{Status: api.StatusBadRequest, Field: reader.Field, Message: reader.Message}
}

// QueryEscape is text escaped for a value of a query string.
func QueryEscape(sandbox *api.Sandbox, text string) string {
	const digits = "0123456789ABCDEF"
	escaped := []byte{}
	for _, char := range []byte(text) {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '-', char == '_', char == '.', char == '~':
			escaped = append(escaped, char)
		case char == ' ':
			escaped = append(escaped, '+')
		default:
			escaped = append(escaped, '%', digits[char>>4], digits[char&15])
		}
	}
	return string(escaped)
}

// Href is path with the query pairs (key, value, key, value…) whose value is
// not "" nor "0".
func Href(sandbox *api.Sandbox, path string, pairs ...string) string {
	query := []string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] == "" || pairs[i+1] == "0" {
			continue
		}
		query = append(query, pairs[i]+"="+QueryEscape(sandbox, pairs[i+1]))
	}
	if len(query) == 0 {
		return path
	}
	return path + "?" + sandbox.Deps.StringsDeps.Join(query, "&")
}
