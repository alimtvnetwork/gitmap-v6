package cmd

import (
	"os"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
)

// dispatchCore routes scan, clone, pull, and status commands.
func dispatchCore(command string) bool {
	if command == constants.CmdScan || command == constants.CmdScanAlias {
		runScan(os.Args[2:])

		return true
	}
	if command == constants.CmdClone || command == constants.CmdCloneAlias {
		runClone(os.Args[2:])

		return true
	}
	if command == constants.CmdPull || command == constants.CmdPullAlias {
		runPull(os.Args[2:])

		return true
	}
	if command == constants.CmdStatus || command == constants.CmdStatusAlias {
		runStatus(os.Args[2:])

		return true
	}
	if command == constants.CmdExec || command == constants.CmdExecAlias {
		runExec(os.Args[2:])

		return true
	}
	if command == constants.CmdHasAnyUpdates || command == constants.CmdHasAnyUpdatesAlias ||
		command == constants.CmdHasAnyChanges || command == constants.CmdHasAnyChangesAlias {
		runHasAnyUpdates(os.Args[2:])

		return true
	}
	if command == constants.CmdHasChange || command == constants.CmdHasChangeAlias {
		runHasChange(os.Args[2:])

		return true
	}
	if command == constants.CmdCloneNext || command == constants.CmdCloneNextAlias {
		runCloneNext(os.Args[2:])

		return true
	}
	if command == constants.CmdAs || command == constants.CmdAsAlias {
		runAs(os.Args[2:])

		return true
	}

	return false
}
