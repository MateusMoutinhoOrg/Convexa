package set_third_party_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpview"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /admin/manage/set-third-party/{id}: the third party is
// rewritten from the form and the browser sent to the list; a form it cannot
// be written from is answered again, under a 400 with the reason above it.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	id := int64(input.Id)
	fields := erpstore.ThirdPartyFields{
		Name: input.Body.Name, Role: input.Body.Role, Document: input.Body.Document, Email: input.Body.Email,
		Phone: input.Body.Phone, Address: input.Body.Address, Notes: input.Body.Notes, Status: input.Body.Status,
	}
	problem := erpstore.Problem{}
	if !erpstore.Failed(sandbox, problem) {
		var err error
		_, problem, err = erpstore.SetThirdParty(sandbox, id, fields)
		if err != nil {
			return err
		}
	}
	if problem.Status == api.StatusNotFound {
		return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-third-parties", erpview.NoticeNotFound))
	}
	if erpstore.Failed(sandbox, problem) {
		return erpview.RenderThirdPartyForm(sandbox, response, props, id, fields, problem)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, erpview.WithNotice(sandbox, "/admin/erp/list-third-parties", erpview.NoticeUpdated))
}
