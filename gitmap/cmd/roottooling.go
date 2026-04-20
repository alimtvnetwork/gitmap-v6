package cmd

import (
	"os"

	"github.com/user/gitmap/constants"
)

// dispatchTooling routes dev tooling and maintenance commands.
func dispatchTooling(command string) bool {
	if command == constants.CmdDesktopSync || command == constants.CmdDesktopSyncAlias {
		checkHelp("desktop-sync", os.Args[2:])
		runDesktopSync()

		return true
	}
	if command == constants.CmdRescan || command == constants.CmdRescanAlias {
		checkHelp("rescan", os.Args[2:])
		runRescan()

		return true
	}
	if command == constants.CmdSetup {
		runSetup(os.Args[2:])

		return true
	}
	if command == constants.CmdDoctor {
		checkHelp("doctor", os.Args[2:])
		runDoctor()

		return true
	}
	if command == constants.CmdLatestBranch || command == constants.CmdLatestBranchAlias {
		runLatestBranch(os.Args[2:])

		return true
	}
	if command == constants.CmdListVersions || command == constants.CmdListVersionsAlias {
		runListVersions(os.Args[2:])

		return true
	}
	if command == constants.CmdListReleases || command == constants.CmdListReleasesAlias {
		runListReleases(os.Args[2:])

		return true
	}
	if command == constants.CmdSEOWrite || command == constants.CmdSEOWriteAlias {
		runSEOWrite(os.Args[2:])

		return true
	}
	if command == constants.CmdGoMod || command == constants.CmdGoModAlias {
		runGoMod(os.Args[2:])

		return true
	}
	if command == constants.CmdCompletion || command == constants.CmdCompletionAlias {
		runCompletion(os.Args[2:])

		return true
	}
	if command == constants.CmdZipGroup || command == constants.CmdZipGroupShort {
		runZipGroup(os.Args[2:])

		return true
	}
	if command == constants.CmdAlias || command == constants.CmdAliasShort {
		runAlias(os.Args[2:])

		return true
	}
	if command == constants.CmdSSH {
		runSSH(os.Args[2:])

		return true
	}
	if command == constants.CmdPrune || command == constants.CmdPruneAlias {
		runPrune(os.Args[2:])

		return true
	}
	if command == constants.CmdTempRelease || command == constants.CmdTempReleaseShort {
		runTempRelease(os.Args[2:])

		return true
	}
	if command == constants.CmdTask || command == constants.CmdTaskAlias {
		runTask(os.Args[2:])

		return true
	}
	if command == constants.CmdEnv || command == constants.CmdEnvAlias {
		runEnv(os.Args[2:])

		return true
	}
	if command == constants.CmdInstall || command == constants.CmdInstallAlias {
		runInstall(os.Args[2:])

		return true
	}
	if command == constants.CmdUninstall || command == constants.CmdUninstallAlias {
		runUninstall(os.Args[2:])

		return true
	}
	if command == constants.CmdSelfInstall {
		runSelfInstall(os.Args[2:])

		return true
	}
	if command == constants.CmdSelfUninstall {
		runSelfUninstall(os.Args[2:])

		return true
	}
	if command == constants.CmdSelfUninstallRunner {
		runSelfUninstallRunner()

		return true
	}
	if command == constants.CmdPending {
		runPending()

		return true
	}
	if command == constants.CmdDoPending || command == constants.CmdDoPendingAlias {
		runDoPending(os.Args[2:])

		return true
	}

	return false
}
