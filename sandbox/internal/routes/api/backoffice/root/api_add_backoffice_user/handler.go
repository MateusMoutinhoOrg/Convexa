package api_add_backoffice_user

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficeusers"
)

// Handle answers POST /api/admin/root/add-backoffice-user. A user
// the body describes well is added and answered under a 201; anything else is
// refused with a 400 carrying the reason.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	parsed, ok := backofficeauth.ParseRole(sandbox, input.Body.Role)
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "role", "role must be root, admin or basic")
	}
	role := int64(parsed)
	user, message, err := backofficeusers.Add(sandbox, backofficeusers.Fields{
		Username: input.Body.Username,
		Email:    input.Body.Email,
		Password: input.Body.Password,
		Role:     role,
	})
	if err != nil {
		return err
	}
	if message != "" {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "", message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusCreated, backofficeapi.UserResponseJSON(sandbox, user))
}
