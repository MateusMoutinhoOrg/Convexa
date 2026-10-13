package api_dashboard

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/dashboard: the overview of every company's
// finances today, its projection looking months ahead (12 when 0).
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.DashboardJSON(sandbox, erpreports.ComputeDashboard(sandbox, book, int64(input.Body.Months))))
}
