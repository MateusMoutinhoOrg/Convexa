package api_add_company

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-company: the company the body
// describes is added and answered under a 201; one it cannot be added from is
// refused, naming the field at fault.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	fields := erpstore.CompanyFields{Name: input.Body.Name, Document: input.Body.Document, Notes: input.Body.Notes, Status: input.Body.Status}
	record, problem, err := erpstore.AddCompany(sandbox, fields)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusCreated, "company", erpapi.CompanyJSON(sandbox, record))
}
