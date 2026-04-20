package cmd

import (
	"os"

	"github.com/user/gitmap/gitutil"
	"github.com/user/gitmap/helptext"
)

// checkHelp prints embedded help and exits if --help or -h is present.
func checkHelp(command string, args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			helptext.Print(command)
			os.Exit(0)
		}
	}
}

// requireOnline checks network connectivity and exits if offline.
func requireOnline() {
	if gitutil.IsOnline() {
		return
	}

	gitutil.PrintOfflineWarning()
	os.Exit(1)
}
