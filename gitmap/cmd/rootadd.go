package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// dispatchAdd routes `gitmap add <subcommand>` calls. Returns true when the
// top-level command was `add` (or its alias), regardless of subcommand
// validity.
func dispatchAdd(command string) bool {
	if command != constants.CmdAdd && command != constants.CmdAddAlias {
		return false
	}
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stderr, constants.UsageAddRoot)
		os.Exit(1)
	}

	sub, rest := os.Args[2], os.Args[3:]
	switch sub {
	case constants.AddSubIgnore, constants.AddSubIgnoreAlias:
		runAddIgnore(rest)
	case constants.AddSubAttributes, constants.AddSubAttributesAlias:
		runAddAttributes(rest)
	case constants.AddSubLFSInstall, constants.AddSubLFSInstallAlias:
		runAddLFSInstall(rest)
	default:
		fmt.Fprintf(os.Stderr, constants.ErrUnknownAddSubcommand, sub)
		fmt.Fprint(os.Stderr, constants.UsageAddRoot)
		os.Exit(1)
	}

	return true
}
