package erppage

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpmoney"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
)

// Every page of Convexa — the backoffice's own and the ERP's — is two
// templates parsed into one set by Deps.PageDeps: assets/backoffice/layout.html,
// which holds the head, the styles, the navigation and the notices, and the
// page's own file, which defines the "content" block the layout leaves open.
// The page's data travels as Page.Body; the layout reads the rest of Page.

// LayoutPath is the template every page is rendered inside.
const LayoutPath = "backoffice/layout.html"

// PartialsPath holds the blocks the pages share: a form's error, its
// buttons, its remove button, the status filter.
const PartialsPath = "backoffice/partials.html"

// Viewer is the signed-in user, as the layout shows them and as a page
// decides what to offer them.
type Viewer struct {
	Id       int64
	Username string
	// Initial is the first letter of Username, shown in the avatar.
	Initial string
	Email   string
	// Role is the name of the user's role: root, admin or basic.
	Role string
	// IsRoot shows what only a root may do: the users, the backups, the
	// settings.
	IsRoot bool
	// CanManage shows what changes a financial record: root and admin.
	CanManage bool
}

// Notice is the outcome of the last action, shown above the page.
type Notice struct {
	Text string
	// Kind is "ok" or "error", and styles the banner.
	Kind string
}

// NavItem is one link of the navigation.
type NavItem struct {
	Key    string
	Label  string
	Href   string
	Active bool
}

// NavGroup is one heading of the navigation and the links under it.
type NavGroup struct {
	Label string
	Items []NavItem
}

// LanguageOption is one language the switcher offers.
type LanguageOption struct {
	Code   string
	Name   string
	Active bool
}

// Page is what layout.html is rendered with.
type Page struct {
	// HtmlLang is the lang attribute of the page.
	HtmlLang string
	// Language is "en" or "pt".
	Language string
	// T translates a text into Language: {{call .T "Save"}}.
	T func(string) string
	// Tf translates a format into Language and fills it:
	// {{call .Tf "%d of %d" .From .To}}.
	Tf func(string, ...any) string
	// Money spells cents the way Language writes an amount: {{call $.Money .Amount}}.
	Money func(int64) string
	// Sign is the class an amount is styled with: "pos", "neg" or "".
	Sign func(int64) string
	// Date spells a YYYYMMDD date the way Language writes one; Month a
	// YYYYMM month; ISO a date the way a date input is filled.
	Date  func(int64) string
	Month func(int64) string
	ISO   func(int64) string
	// Input spells cents the way an amount input is filled.
	Input func(int64) string
	// Today is today's date, YYYYMMDD.
	Today  int64
	Title  string
	Viewer Viewer
	// SignedIn shows the navigation; the sign-in page has none.
	SignedIn  bool
	Nav       []NavGroup
	Languages []LanguageOption
	Notice    Notice
	// Refresh reloads the page every few seconds, while a job runs.
	Refresh bool
	// Path is where the language switcher returns to.
	Path string
	// Body is the data of the page's own template.
	Body any
}

// Language is the language a page is shown in: the one the signed-in user
// chose, else the one the request asked for (?lang= or the cookie it left),
// else the browser's, else English.
func Language(sandbox *api.Sandbox, props *routeprops.RouteProps) string {
	if props.User != nil {
		if lang := erpi18n.Normalize(sandbox, props.User.Language); lang != "" {
			return lang
		}
	}
	if props.Language != "" {
		return props.Language
	}
	if props.AcceptLanguage != "" {
		return props.AcceptLanguage
	}
	return erpi18n.English
}

// ApiLanguage is the language the api answers a message in: the one the
// call asked for (?lang= or Accept-Language), else the one its token's user
// chose, else English. A caller of the api says what it wants on every call,
// so the call wins over the user's choice where a page lets the choice win.
func ApiLanguage(sandbox *api.Sandbox, props *routeprops.RouteProps) string {
	if props.Language != "" {
		return props.Language
	}
	if props.AcceptLanguage != "" {
		return props.AcceptLanguage
	}
	if props.User != nil {
		if lang := erpi18n.Normalize(sandbox, props.User.Language); lang != "" {
			return lang
		}
	}
	return erpi18n.English
}

// initialOf is the first letter of name, "?" for an empty one.
func initialOf(sandbox *api.Sandbox, name string) string {
	for _, letter := range name {
		return string(letter)
	}
	return "?"
}

// ViewerOf is the Viewer of user; the zero Viewer for nil.
func ViewerOf(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord) Viewer {
	if user == nil {
		return Viewer{}
	}
	role := backofficeauth.Role(user.Role)
	return Viewer{
		Id:        user.Id,
		Username:  user.Username,
		Initial:   initialOf(sandbox, user.Username),
		Email:     user.Email,
		Role:      backofficeauth.RoleName(sandbox, role),
		IsRoot:    role == backofficeauth.RoleRoot,
		CanManage: backofficeauth.CanManage(sandbox, role),
	}
}

// CanManage tells whether the user of props may change the financial records.
func CanManage(sandbox *api.Sandbox, props *routeprops.RouteProps) bool {
	return props.User != nil && backofficeauth.CanManage(sandbox, backofficeauth.Role(props.User.Role))
}

// nav is the navigation viewer sees in lang, the item keyed active marked.
func nav(sandbox *api.Sandbox, lang string, viewer Viewer, active string) []NavGroup {
	t := erpi18n.Translator(sandbox, lang)
	item := func(key string, label string, href string) NavItem {
		return NavItem{Key: key, Label: t(label), Href: href, Active: key == active}
	}
	admin := []NavItem{
		item("users", "Users", "/admin/list-backoffice-users"),
		item("tokens", "API tokens", "/admin/list-backoffice-api-tokens"),
	}
	if viewer.IsRoot {
		admin = append(admin,
			item("backups", "Backups", "/admin/root/list-backups"),
			item("settings", "Settings", "/admin/root/settings"),
		)
	}
	admin = append(admin, item("home", "My account", "/admin/home"))
	return []NavGroup{
		{Label: t("Overview"), Items: []NavItem{
			item("dashboard", "Dashboard", "/admin/erp/dashboard"),
		}},
		{Label: t("Finance"), Items: []NavItem{
			item("transactions", "Transactions", "/admin/erp/list-transactions"),
			item("payables", "Payables & receivables", "/admin/erp/payables-receivables"),
			item("cards", "Credit cards", "/admin/erp/list-credit-cards"),
			item("budgets", "Budgets", "/admin/erp/list-budgets"),
			item("recurrences", "Recurrences", "/admin/erp/list-recurrences"),
		}},
		{Label: t("Reports"), Items: []NavItem{
			item("cash-flow", "Cash flow", "/admin/erp/cash-flow"),
			item("projection", "Projected balance", "/admin/erp/projection"),
			item("company-report", "By company", "/admin/erp/company-report"),
			item("account-report", "By account", "/admin/erp/account-report"),
			item("budget-report", "Budget vs actual", "/admin/erp/budget-report"),
		}},
		{Label: t("Registrations"), Items: []NavItem{
			item("companies", "Companies", "/admin/erp/list-companies"),
			item("accounts", "Accounts", "/admin/erp/list-accounts"),
			item("categories", "Categories", "/admin/erp/list-categories"),
			item("third-parties", "Third parties", "/admin/erp/list-third-parties"),
		}},
		{Label: t("Administration"), Items: admin},
	}
}

// languages are the options of the switcher, lang marked.
func languages(sandbox *api.Sandbox, lang string) []LanguageOption {
	options := []LanguageOption{}
	for _, code := range erpi18n.Languages(sandbox) {
		options = append(options, LanguageOption{Code: code, Name: erpi18n.Name(sandbox, code), Active: code == lang})
	}
	return options
}

// NewPage is the Page of a request with props: titled title, translated, the
// navigation item active marked, notice above it, body for the page's
// template.
func NewPage(sandbox *api.Sandbox, props *routeprops.RouteProps, title string, active string, notice Notice, body any) Page {
	lang := Language(sandbox, props)
	viewer := ViewerOf(sandbox, props.User)
	page := Page{
		HtmlLang:  erpi18n.HtmlLang(sandbox, lang),
		Language:  lang,
		T:         erpi18n.Translator(sandbox, lang),
		Tf:        erpi18n.Formatter(sandbox, lang),
		Money:     func(cents int64) string { return erpmoney.Format(sandbox, lang, cents) },
		Sign:      func(cents int64) string { return signClass(sandbox, cents) },
		Date:      func(date int64) string { return erpdate.Format(sandbox, lang, date) },
		Month:     func(month int64) string { return erpdate.FormatMonth(sandbox, lang, month) },
		ISO:       func(date int64) string { return erpdate.ISO(sandbox, date) },
		Input:     func(cents int64) string { return erpmoney.FormatInput(sandbox, lang, cents) },
		Today:     erpdate.Today(sandbox),
		Title:     erpi18n.Tr(sandbox, lang, title),
		Viewer:    viewer,
		SignedIn:  props.User != nil,
		Languages: languages(sandbox, lang),
		Notice:    Notice{Text: erpi18n.Tr(sandbox, lang, notice.Text), Kind: notice.Kind},
		Path:      props.Path,
		Body:      body,
	}
	if page.SignedIn {
		page.Nav = nav(sandbox, lang, viewer, active)
	}
	return page
}

// signClass is the class an amount of cents is styled with.
func signClass(sandbox *api.Sandbox, cents int64) string {
	if cents > 0 {
		return "pos"
	}
	if cents < 0 {
		return "neg"
	}
	return ""
}

// Write answers page under status, rendered from layout.html and the page
// template at template (a path under assets/, e.g. "backoffice/home.html").
func Write(sandbox *api.Sandbox, response *serverdeps.Response, status int, template string, page Page) error {
	content, err := sandbox.Deps.PageDeps.RenderHTMLPage([]string{LayoutPath, PartialsPath, template}, "layout", page)
	if err != nil {
		return err
	}
	response.SetHeader("Content-Type", "text/html; charset=utf-8")
	response.SetHeader("Cache-Control", "no-store")
	response.SetStatus(status)
	return response.Write(content)
}

// Render answers, under status, the page template at template for the
// request with props: NewPage then Write.
func Render(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, template string, title string, active string, notice Notice, body any) error {
	return Write(sandbox, response, status, template, NewPage(sandbox, props, title, active, notice, body))
}

// Ok is a notice of success with text.
func Ok(sandbox *api.Sandbox, text string) Notice {
	return Notice{Text: text, Kind: "ok"}
}

// Error is a notice of failure with text, Notice{} for "".
func Error(sandbox *api.Sandbox, text string) Notice {
	if text == "" {
		return Notice{}
	}
	return Notice{Text: text, Kind: "error"}
}
