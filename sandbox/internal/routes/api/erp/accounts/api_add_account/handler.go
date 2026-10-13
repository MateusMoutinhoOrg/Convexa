package api_add_account

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/manage/add-account: the account the body
// describes is added and answered under a 201; one it cannot be added from is
// refused, naming the field at fault.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	openingDate, err := erpapi.Date(sandbox, props, "opening-date", input.Body.OpeningDate)
	if err != nil {
		return err
	}
	fields := erpstore.AccountFields{
		Name: input.Body.Name, Kind: input.Body.Kind, Institution: input.Body.Institution, Notes: input.Body.Notes, Status: input.Body.Status,
		CreditLimit: int64(input.Body.CreditLimit), ClosingDay: int64(input.Body.ClosingDay), DueDay: int64(input.Body.DueDay),
		PaymentAccount: int64(input.Body.PaymentAccount), OpeningBalance: int64(input.Body.OpeningBalance),
		OpeningCompany: int64(input.Body.OpeningCompany), OpeningDate: openingDate,
	}
	record, problem, err := erpstore.AddAccount(sandbox, fields, props.User.Id)
	if err != nil {
		return err
	}
	if erpstore.Failed(sandbox, problem) {
		return erpapi.Refuse(sandbox, props, problem)
	}
	return erpapi.One(sandbox, response, api.StatusCreated, "account", erpapi.AccountJSON(sandbox, record, 0, 0, nil))
}
