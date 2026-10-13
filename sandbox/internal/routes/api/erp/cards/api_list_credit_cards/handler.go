package api_list_credit_cards

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpapi"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpreports"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/erp/list-credit-cards: {"cards": [...]}, every
// credit card with its debt, what is available of its limit, its open invoice
// and the oldest one not paid yet.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	book, err := erpreports.Load(sandbox)
	if err != nil {
		return err
	}
	cards := erpapi.Array(sandbox)
	for _, summary := range erpreports.ComputeCardSummaries(sandbox, book) {
		cards.AddItemToArray(erpapi.CardSummaryJSON(sandbox, summary))
	}
	document := erpapi.Object(sandbox)
	document.AddItemToObject("cards", cards)
	return erpapi.Write(sandbox, response, api.StatusOK, document)
}
