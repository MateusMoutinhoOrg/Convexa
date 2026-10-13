package optimize_backup_storage_form

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/server/backoffice/backofficesnapshots"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/snapshots"
)

// Handle answers POST /admin/root/optimize-backup-storage: every
// stored content no snapshot holds anymore starts being removed in the
// background, and the browser is sent to the list at once. Another backup job
// running refuses it with a notice.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	notice := backofficesnapshots.NoticeOptimizing
	if !snapshots.StartOptimize(sandbox) {
		notice = backofficesnapshots.NoticeBusy
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, backofficesnapshots.ListLocation(sandbox, notice))
}
