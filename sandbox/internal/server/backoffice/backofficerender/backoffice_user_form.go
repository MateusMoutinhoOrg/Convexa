package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeusers"
)

// BackofficeUserFormPage is what backoffice/backoffice_user_form.html is
// rendered with: the form that adds a backoffice user, or the one that edits
// one.
type BackofficeUserFormPage struct {
	// Title heads the page.
	Title string
	// Action is where the form posts.
	Action string
	// Submit labels the submit button.
	Submit string
	// Error is shown above the form, "" for none.
	Error    string
	Username string
	Email    string
	Roles    []RoleOption
	// PasswordRequired is true when adding: an edit keeps the current
	// password when the field is left blank.
	PasswordRequired bool
}

// RoleOption is one role the form offers.
type RoleOption struct {
	Value    int64
	Name     string
	Selected bool
}

// roleOptions are every role, the one of fields selected.
func roleOptions(sandbox *api.Sandbox, fields backofficeusers.Fields) []RoleOption {
	options := []RoleOption{}
	for _, role := range backofficeauth.Roles(sandbox) {
		options = append(options, RoleOption{
			Value:    int64(role),
			Name:     backofficeauth.RoleName(sandbox, role),
			Selected: int64(role) == fields.Role,
		})
	}
	return options
}

// RenderAddUserPage answers, under status, the form that adds a
// backoffice user, filled with fields — never their password — and with
// message above it.
func RenderAddUserPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, fields backofficeusers.Fields, message string) error {
	return render(sandbox, response, status, props, "backoffice/backoffice_user_form.html", "Add user", "users", Notice{}, BackofficeUserFormPage{
		Title:            tr(sandbox, props, "Add user"),
		Action:           "/admin/root/add-backoffice-user",
		Submit:           tr(sandbox, props, "Add user"),
		Error:            tr(sandbox, props, message),
		Username:         fields.Username,
		Email:            fields.Email,
		Roles:            roleOptions(sandbox, fields),
		PasswordRequired: true,
	})
}

// RenderSetUserPage answers, under status, the form that edits the
// backoffice user with id id, filled with fields — never their password — and
// with message above it.
func RenderSetUserPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, id int64, fields backofficeusers.Fields, message string) error {
	return render(sandbox, response, status, props, "backoffice/backoffice_user_form.html", "Edit user", "users", Notice{}, BackofficeUserFormPage{
		Title:    trf(sandbox, props, "Edit user #%d", id),
		Action:   "/admin/root/set-backoffice-user/" + sandbox.Deps.StringsDeps.FormatInt(id, 10),
		Submit:   tr(sandbox, props, "Save changes"),
		Error:    tr(sandbox, props, message),
		Username: fields.Username,
		Email:    fields.Email,
		Roles:    roleOptions(sandbox, fields),
	})
}
