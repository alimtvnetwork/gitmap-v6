package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/verbose"
)

// runRevertRunner is the hidden command that performs the revert build.
func runRevertRunner() {
	repoPath := constants.RepoPath
	if len(repoPath) == 0 {
		fmt.Fprint(os.Stderr, constants.ErrNoRepoPath)
		os.Exit(1)
	}

	initRunnerVerbose()
	fmt.Printf(constants.MsgRevertStarting)
	executeRevert(repoPath)
}

// executeRevert writes a temp PS1 script and runs it.
func executeRevert(repoPath string) {
	scriptPath, err := writeRevertScript(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrRevertFailed, err)
		os.Exit(1)
	}
	defer os.Remove(scriptPath)

	log := verbose.Get()
	if log != nil {
		log.Log(constants.RevertScriptLogExec, scriptPath)
	}

	runRevertPS(scriptPath)
}

// writeRevertScript creates a temporary PowerShell script for revert build.
func writeRevertScript(repoPath string) (string, error) {
	runPS1 := filepath.Join(repoPath, "run.ps1")
	script := buildRevertScript(repoPath, runPS1)

	return writeScriptToTemp(script)
}

// buildRevertScript generates the PowerShell script content for revert.
func buildRevertScript(repoPath, runPS1 string) string {
	return fmt.Sprintf(constants.RevertPSHeader, repoPath) +
		fmt.Sprintf(constants.RevertPSBuild, runPS1) +
		constants.RevertPSPostActions
}

// runRevertPS executes the PowerShell script with output piped to terminal.
func runRevertPS(scriptPath string) {
	cmd := exec.Command(constants.PSBin, constants.PSExecPolicy, constants.PSBypass,
		constants.PSNoProfile, constants.PSNoLogo, constants.PSFile, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	logRevertResult(err)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrRevertFailed, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgRevertDone)
}

// logRevertResult logs the revert script exit status if verbose is active.
func logRevertResult(err error) {
	log := verbose.Get()
	if log != nil {
		log.Log(constants.RevertScriptLogExit, err)
	}
}
