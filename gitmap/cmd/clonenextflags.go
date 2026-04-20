package cmd

import (
	"flag"

	"github.com/user/gitmap/constants"
)

// parseCloneNextFlags parses flags for the clone-next command.
func parseCloneNextFlags(args []string) (versionArg string, deleteFlag, keepFlag, noDesktop, createRemote bool, sshKeyName string, verbose bool) {
	fs := flag.NewFlagSet(constants.CmdCloneNext, flag.ExitOnError)
	delFlag := fs.Bool("delete", false, constants.FlagDescCloneNextDelete)
	kpFlag := fs.Bool("keep", false, constants.FlagDescCloneNextKeep)
	noDesk := fs.Bool("no-desktop", false, constants.FlagDescCloneNextNoDesktop)
	createRem := fs.Bool("create-remote", false, constants.FlagDescCloneNextCreateRemote)
	sshKey := fs.String("ssh-key", "", "SSH key name for clone")
	fs.StringVar(sshKey, "K", "", "SSH key name (short)")
	verboseFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	fs.Parse(args)

	if fs.NArg() > 0 {
		versionArg = fs.Arg(0)
	}

	return versionArg, *delFlag, *kpFlag, *noDesk, *createRem, *sshKey, *verboseFlag
}
