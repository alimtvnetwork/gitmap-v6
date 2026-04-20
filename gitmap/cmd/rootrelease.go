package cmd

import (
	"os"

	"github.com/user/gitmap/constants"
)

// dispatchRelease routes release-related commands.
func dispatchRelease(command string) bool {
	if command == constants.CmdRelease || command == constants.CmdReleaseAlias {
		runRelease(os.Args[2:])

		return true
	}
	if command == constants.CmdReleaseSelf || command == constants.CmdReleaseSelfAlias || command == constants.CmdReleaseSelfAlias2 {
		runReleaseSelf(os.Args[2:])

		return true
	}
	if command == constants.CmdReleaseBranch || command == constants.CmdReleaseBranchAlias {
		runReleaseBranch(os.Args[2:])

		return true
	}
	if command == constants.CmdReleasePending || command == constants.CmdReleasePendingAlias {
		runReleasePending(os.Args[2:])

		return true
	}
	if command == constants.CmdChangelog || command == constants.CmdChangelogAlias {
		runChangelog(os.Args[2:])

		return true
	}
	if command == constants.CmdChangelogMD {
		runChangelog([]string{constants.FlagOpenValue})

		return true
	}
	if command == constants.CmdClearReleaseJSON || command == constants.CmdClearReleaseJSONAlias {
		runClearReleaseJSON(os.Args[2:])

		return true
	}
	if command == constants.CmdChangelogGen || command == constants.CmdChangelogGenAlias {
		runChangelogGen(os.Args[2:])

		return true
	}
	if command == constants.CmdReleaseAlias || command == constants.CmdReleaseAliasShort {
		runReleaseAlias(os.Args[2:], false)

		return true
	}
	if command == constants.CmdReleaseAliasPull || command == constants.CmdReleaseAliasPullShort {
		runReleaseAlias(os.Args[2:], true)

		return true
	}

	return false
}
