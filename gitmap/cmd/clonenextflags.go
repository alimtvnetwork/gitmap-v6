package cmd

import (
	"flag"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
)

// CloneNextFlags bundles every parsed flag from the clone-next command so
// the dispatcher in runCloneNext can branch on batch vs single mode without
// a 9-arg return list.
type CloneNextFlags struct {
	VersionArg   string
	Delete       bool
	Keep         bool
	NoDesktop    bool
	CreateRemote bool
	SSHKeyName   string
	Verbose      bool
	CSVPath      string
	All          bool
}

// parseCloneNextFlags parses flags for the clone-next command.
func parseCloneNextFlags(args []string) CloneNextFlags {
	fs := flag.NewFlagSet(constants.CmdCloneNext, flag.ExitOnError)
	delFlag := fs.Bool("delete", false, constants.FlagDescCloneNextDelete)
	kpFlag := fs.Bool("keep", false, constants.FlagDescCloneNextKeep)
	noDesk := fs.Bool("no-desktop", false, constants.FlagDescCloneNextNoDesktop)
	createRem := fs.Bool("create-remote", false, constants.FlagDescCloneNextCreateRemote)
	sshKey := fs.String("ssh-key", "", "SSH key name for clone")
	fs.StringVar(sshKey, "K", "", "SSH key name (short)")
	verboseFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	csvPath := fs.String("csv", "", constants.FlagDescCloneNextCSV)
	allFlag := fs.Bool("all", false, constants.FlagDescCloneNextAll)
	fs.Parse(args)

	out := CloneNextFlags{
		Delete:       *delFlag,
		Keep:         *kpFlag,
		NoDesktop:    *noDesk,
		CreateRemote: *createRem,
		SSHKeyName:   *sshKey,
		Verbose:      *verboseFlag,
		CSVPath:      *csvPath,
		All:          *allFlag,
	}
	if fs.NArg() > 0 {
		out.VersionArg = fs.Arg(0)
	}

	return out
}
