package api_projection

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/projection: the projected balance months
// ahead (12 when 0), with and without the budgets, month by month — and every
// event to come, when events is true.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	projection := erpreports.ComputeProjection(sandbox, book, erpreports.ProjectionQuery{
		Months: int64(input.Body.Months), Company: int64(input.Body.Company), Account: int64(input.Body.Account),
	})
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.ProjectionJSON(sandbox, projection, input.Body.Events))
}
