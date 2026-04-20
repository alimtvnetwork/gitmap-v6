package helptext

import (
	"embed"
	"fmt"
	"os"
)

//go:embed *.md
var files embed.FS

// Print reads and prints the help file for the given command.
func Print(command string) {
	data, err := files.ReadFile(command + ".md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "No help available for '%s'\n", command)
		os.Exit(1)
	}
	fmt.Print(string(data))
}
