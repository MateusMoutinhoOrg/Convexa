package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeapitokens"
)

// BackofficeApiTokensPage is what backoffice/backoffice_api_tokens.html is
// rendered with.
type BackofficeApiTokensPage struct {
	// Created is the token just created, shown once above the list; its
	// Token is "" for none.
	Created CreatedToken
	Tokens  []BackofficeApiTokenRow
	// ShowOwner adds the owner column, for a root, who sees every user's
	// tokens.
	ShowOwner bool
}

// CreatedToken is a token the list page shows in full, right after it was
// created: the one time it is ever shown.
type CreatedToken struct {
	Name  string
	Token string
}

// BackofficeApiTokenRow is one token of the list.
type BackofficeApiTokenRow struct {
	Id     string
	Name   string
	Prefix string
	Owner  string
	// IsOwn marks a token of the signed-in user, on a root's list.
	IsOwn   bool
	Created string
	// Expires is the instant the token expires, or "Never".
	Expires   string
	IsExpired bool
	// Ips are the ips the token is accepted from, or "Any".
	Ips string
	// LastUsed is the instant and the ip of the token's last request, or
	// "Never".
	LastUsed string
}

// apiTokenNoticeOf is how the list page words a backofficeapitokens notice code;
// an unknown code shows nothing.
func apiTokenNoticeOf(sandbox *api.Sandbox, code string) Notice {
	switch code {
	case backofficeapitokens.NoticeRevoked:
		return Notice{Text: "Token revoked. It no longer authenticates any request.", Kind: "ok"}
	case backofficeapitokens.NoticeNotFound:
		return Notice{Text: "That token no longer exists.", Kind: "error"}
	}
	return Notice{}
}

// instantOf is the instant seconds as the pages of the request with props
// spell it.
func instantOf(sandbox *api.Sandbox, props *routeprops.RouteProps, seconds int64) string {
	return erpdate.FormatInstant(sandbox, erppage.Language(sandbox, props), seconds)
}

// apiTokenRow is the row of listed on the list page of the request with
// props, at now.
func apiTokenRow(sandbox *api.Sandbox, props *routeprops.RouteProps, listed backofficeapitokens.ApiTokenRow, now int64) BackofficeApiTokenRow {
	user := props.User
	token := listed.Token
	row := BackofficeApiTokenRow{
		Id:        sandbox.Deps.StringsDeps.FormatInt(token.Id, 10),
		Name:      token.Name,
		Prefix:    token.Prefix,
		Owner:     listed.Owner.Username,
		IsOwn:     listed.Owner.Id == user.Id,
		Created:   instantOf(sandbox, props, token.CreatedAt),
		Expires:   tr(sandbox, props, "Never"),
		IsExpired: backofficeapitokens.Expired(sandbox, token, now),
		Ips:       tr(sandbox, props, "Any"),
		LastUsed:  tr(sandbox, props, "Never"),
	}
	if token.ExpiresAt != 0 {
		row.Expires = instantOf(sandbox, props, token.ExpiresAt)
	}
	ips := backofficeapitokens.IpList(sandbox, token)
	if len(ips) > 0 {
		row.Ips = sandbox.Deps.StringsDeps.Join(ips, ", ")
	}
	if token.LastUsedAt != 0 {
		row.LastUsed = trf(sandbox, props, "%s from %s", instantOf(sandbox, props, token.LastUsedAt), token.LastUsedIp)
	}
	return row
}

// RenderApiTokensPage answers, under status, the API token list page for
// user: listed, with the notice code notice above it, and created shown in
// full when its Token is not "".
func RenderApiTokensPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, props *routeprops.RouteProps, listed []backofficeapitokens.ApiTokenRow, notice string, created CreatedToken) error {
	now := sandbox.Deps.StdDeps.Now() / 1_000_000_000
	rows := []BackofficeApiTokenRow{}
	for _, item := range listed {
		rows = append(rows, apiTokenRow(sandbox, props, item, now))
	}

	viewer := erppage.ViewerOf(sandbox, props.User)
	return render(sandbox, response, status, props, "backoffice/backoffice_api_tokens.html", "API tokens", "tokens", apiTokenNoticeOf(sandbox, notice), BackofficeApiTokensPage{
		Created:   created,
		Tokens:    rows,
		ShowOwner: viewer.IsRoot,
	})
}
