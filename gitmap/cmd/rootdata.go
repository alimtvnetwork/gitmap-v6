package cmd

import (
	"os"

	"github.com/user/gitmap/constants"
)

// dispatchData routes data management, history, profiles, and TUI commands.
func dispatchData(command string) bool {
	if command == constants.CmdList || command == constants.CmdListAlias {
		runList(os.Args[2:])

		return true
	}
	if command == constants.CmdGroup || command == constants.CmdGroupAlias {
		runGroup(os.Args[2:])

		return true
	}
	if command == constants.CmdMultiGroup || command == constants.CmdMultiGroupAlias {
		runMultiGroup(os.Args[2:])

		return true
	}
	if command == constants.CmdHistory || command == constants.CmdHistoryAlias {
		runHistory(os.Args[2:])

		return true
	}
	if command == constants.CmdHistoryReset || command == constants.CmdHistoryResetAlias {
		runHistoryReset(os.Args[2:])

		return true
	}
	if command == constants.CmdStats || command == constants.CmdStatsAlias {
		runStats(os.Args[2:])

		return true
	}
	if command == constants.CmdBookmark || command == constants.CmdBookmarkAlias {
		runBookmark(os.Args[2:])

		return true
	}
	if command == constants.CmdExport || command == constants.CmdExportAlias {
		runExport(os.Args[2:])

		return true
	}
	if command == constants.CmdImport || command == constants.CmdImportAlias {
		runImport(os.Args[2:])

		return true
	}
	if command == constants.CmdProfile || command == constants.CmdProfileAlias {
		runProfile(os.Args[2:])

		return true
	}
	if command == constants.CmdDiffProfiles || command == constants.CmdDiffProfilesAlias {
		runDiffProfiles(os.Args[2:])

		return true
	}
	if command == constants.CmdCD || command == constants.CmdCDAlias {
		runCD(os.Args[2:])

		return true
	}
	if command == constants.CmdWatch || command == constants.CmdWatchAlias {
		runWatch(os.Args[2:])

		return true
	}
	if command == constants.CmdInteractive || command == constants.CmdInteractiveAlias {
		runInteractive()

		return true
	}
	if command == constants.CmdDBReset {
		runDBReset(os.Args[2:])

		return true
	}
	if command == constants.CmdReset {
		runReset(os.Args[2:])

		return true
	}
	if command == constants.CmdDBMigrate || command == constants.CmdDBMigrateAlias {
		runDBMigrate(os.Args[2:])

		return true
	}
	if command == constants.CmdAmend || command == constants.CmdAmendAlias {
		runAmend(os.Args[2:])

		return true
	}
	if command == constants.CmdAmendList || command == constants.CmdAmendListAlias {
		runAmendList(os.Args[2:])

		return true
	}
	if command == constants.CmdDashboard || command == constants.CmdDashboardAlias {
		runDashboard(os.Args[2:])

		return true
	}
	if command == constants.CmdVersionHistory || command == constants.CmdVersionHistoryAlias {
		runVersionHistory(os.Args[2:])

		return true
	}

	return false
}
