package erpapi

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erppage"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
)

// Fail refuses an api call under status, field naming the input at fault
// ("" for none), with message — written in English — translated into the
// language the call asked for. The JSON shape stays {"error", "field"}
// whatever the language.
func Fail(sandbox *api.Sandbox, props *routeprops.RouteProps, status int, field string, message string) error {
	return sandbox.Deps.OpinionatedAgnosServer.Fail(status, field, erpi18n.Tr(sandbox, erppage.ApiLanguage(sandbox, props), message))
}

// Failf is Fail with a message format filled with args after translating it.
func Failf(sandbox *api.Sandbox, props *routeprops.RouteProps, status int, field string, format string, args ...any) error {
	return sandbox.Deps.OpinionatedAgnosServer.Fail(status, field, erpi18n.Trf(sandbox, erppage.ApiLanguage(sandbox, props), format, args...))
}
