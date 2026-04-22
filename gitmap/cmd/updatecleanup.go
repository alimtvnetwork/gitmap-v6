package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// runUpdateCleanup handles the "update-cleanup" subcommand.
//
// Output is structured into two layers:
//  1. A per-artifact line for every candidate, of the form
//     "<symbol> [<kind>] <path> — <status>: <reason>".
//  2. A summary table grouped by status with totals.
//
// Status codes ("removed", "locked", "missing", "skipped-active", ...) are
// stable so logs are grep-able and CI consumers can parse them.
func runUpdateCleanup() {
	fmt.Println(constants.MsgUpdateCleanStart)

	ctx := loadUpdateCleanupContext()
	report := newCleanupReport()
	cleanupTempArtifacts(ctx, report)
	cleanupBackupArtifacts(ctx, report)
	cleanupDriveRootShim(ctx, report)
	cleanupCloneSwapDirs(ctx, report)
	report.printSummary()
}
