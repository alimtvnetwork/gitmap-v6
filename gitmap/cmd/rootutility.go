package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/helptext"
)

// isFlagToken returns true when arg looks like a CLI flag (-x or --xx).
func isFlagToken(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

// dispatchUtility routes setup, update, doctor, and other utility commands.
func dispatchUtility(command string) bool {
	if command == constants.CmdUpdate {
		checkHelp("update", os.Args[2:])
		runUpdate()

		return true
	}
	if command == constants.CmdUpdateRunner {
		runUpdateRunner()

		return true
	}
	if command == constants.CmdUpdateCleanup {
		runUpdateCleanup()

		return true
	}
	if command == constants.CmdInstalledDir || command == constants.CmdInstalledDirAlias {
		checkHelp("installed-dir", os.Args[2:])
		runInstalledDir()

		return true
	}
	if command == constants.CmdRevert {
		runRevert(os.Args[2:])

		return true
	}
	if command == constants.CmdRevertRunner {
		runRevertRunner()

		return true
	}
	if command == constants.CmdVersion || command == constants.CmdVersionAlias {
		checkHelp("version", os.Args[2:])
		fmt.Printf(constants.MsgVersionFmt, constants.Version)

		return true
	}
	if command == constants.CmdHelp {
		if len(os.Args) >= 3 && !isFlagToken(os.Args[2]) {
			helptext.Print(os.Args[2])

			return true
		}
		if hasFlag(constants.FlagGroups) {
			printHelpGroups()

			return true
		}
		if hasFlag(constants.FlagCompact) {
			printUsageCompact()

			return true
		}
		printUsage()

		return true
	}
	if command == constants.CmdDocs || command == constants.CmdDocsAlias {
		runDocs(os.Args[2:])

		return true
	}
	if command == constants.CmdHelpDashboard || command == constants.CmdHelpDashboardAlias {
		runHelpDashboard(os.Args[2:])

		return true
	}
	if command == constants.CmdLLMDocs || command == constants.CmdLLMDocsAlias {
		runLLMDocs(os.Args[2:])

		return true
	}
	if command == constants.CmdSetSourceRepo {
		runSetSourceRepo()

		return true
	}
	if command == constants.CmdSf {
		runSf(os.Args[2:])

		return true
	}
	if command == constants.CmdProbe {
		runProbe(os.Args[2:])

		return true
	}
	if command == constants.CmdFindNext || command == constants.CmdFindNextAlias {
		runFindNext(os.Args[2:])

		return true
	}

	return false
}
