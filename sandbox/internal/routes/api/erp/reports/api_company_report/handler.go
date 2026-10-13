package api_company_report

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/company-report: what the body's company —
// every one when 0 — did between from and to (this month when blank).
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	from, err := erpapi.Date(sandbox, props, "from", input.Body.From)
	if err != nil {
		return err
	}
	to, err := erpapi.Date(sandbox, props, "to", input.Body.To)
	if err != nil {
		return err
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	if from == 0 {
		from = erpdate.MonthStart(sandbox, book.Today)
	}
	if to == 0 || to < from {
		to = erpdate.MonthEnd(sandbox, from)
	}
	report := erpreports.ComputeCompanyReport(sandbox, book, int64(input.Body.Company), from, to)
	return erpapi.Write(sandbox, response, api.StatusOK, erpapi.CompanyReportJSON(sandbox, book, report))
}
