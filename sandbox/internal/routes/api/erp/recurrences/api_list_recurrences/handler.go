package api_list_recurrences

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpstore"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-recurrences: {"recurrences": [...]},
// every recurrence the body's filters keep, with its next occurrence.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	recurrences, err := erpstore.Recurrences(sandbox, erpstore.RecurrenceQuery{
		Search: input.Body.Search, Company: int64(input.Body.Company), Account: int64(input.Body.Account),
		Type: input.Body.Type, Status: input.Body.Status,
	})
	if err != nil {
		return err
	}
	list := erpapi.Array(sandbox)
	for _, recurrence := range recurrences {
		list.AddItemToArray(erpapi.RecurrenceJSON(sandbox, recurrence, erpstore.NextOccurrence(sandbox, recurrence, erpdate.AddDays(sandbox, erpdate.Today(sandbox), -1))))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("recurrences", list)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
