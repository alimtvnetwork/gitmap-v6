package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/verbose"
)

// executeUpdate writes a temp script and runs it.
// On Windows it uses PowerShell; on Linux/macOS it uses run.sh directly.
func executeUpdate(repoPath string) {
	if runtime.GOOS == "windows" {
		executeUpdateWindows(repoPath)

		return
	}

	executeUpdateUnix(repoPath)
}

// executeUpdateWindows writes a temp PS1 script and runs it.
func executeUpdateWindows(repoPath string) {
	scriptPath, err := writeUpdateScript(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateFailed, err)
		os.Exit(1)
	}
	defer os.Remove(scriptPath)

	log := verbose.Get()
	if log != nil {
		log.Log(constants.UpdateScriptLogExec, scriptPath)
	}

	runUpdateScript(scriptPath)
}

// executeUpdateUnix runs run.sh --update with the install path as deploy target.
func executeUpdateUnix(repoPath string) {
	runSH := filepath.Join(repoPath, "run.sh")

	if _, err := os.Stat(runSH); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateNoRunSH, runSH)
		os.Exit(1)
	}

	log := verbose.Get()
	if log != nil {
		log.Log(constants.UpdateScriptLogExec, runSH)
	}

	// Resolve the active binary's installed directory.
	installDir := resolveInstalledDir()
	fmt.Printf(constants.MsgUpdateInstallDir, installDir)

	// Build args: --update --deploy-path <parent-of-install-dir>
	// run.sh deploys into <deploy-path>/gitmap/, so if the binary is at
	// /home/user/.local/bin/gitmap, deploy-path is /home/user/.local/bin
	// BUT run.sh puts it in <deploy-path>/gitmap/gitmap, which is different.
	// Instead, we just run run.sh --update and let it sync to the PATH binary
	// (lines 601-618 of run.sh already handle this).
	args := []string{runSH, "--update"}

	cmd := exec.Command("bash", args...)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	logScriptResult(err)

	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateFailed, err)
		os.Exit(1)
	}
}

// resolveInstalledDir returns the directory where the active gitmap binary lives.
func resolveInstalledDir() string {
	// First try: which gitmap on PATH
	path, err := exec.LookPath("gitmap")
	if err == nil {
		resolved, evalErr := filepath.EvalSymlinks(path)
		if evalErr == nil {
			return filepath.Dir(resolved)
		}

		return filepath.Dir(path)
	}

	// Fallback: current executable's directory
	selfPath, err := os.Executable()
	if err != nil {
		return ""
	}

	resolved, err := filepath.EvalSymlinks(selfPath)
	if err != nil {
		return filepath.Dir(selfPath)
	}

	return filepath.Dir(resolved)
}

// writeUpdateScript creates a temporary PowerShell script for self-update.
// Writes with UTF-8 BOM so PowerShell correctly handles Unicode characters.
func writeUpdateScript(repoPath string) (string, error) {
	runPS1 := filepath.Join(repoPath, "run.ps1")
	script := buildUpdateScript(repoPath, runPS1)

	return writeScriptToTemp(script)
}

// writeScriptToTemp writes script content to a temp file with UTF-8 BOM.
func writeScriptToTemp(script string) (string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), constants.UpdateScriptGlob)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := tmpFile.Write(bom); err != nil {
		return "", err
	}
	if _, err := tmpFile.WriteString(script); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// buildUpdateScript generates the PowerShell script content.
func buildUpdateScript(repoPath, runPS1 string) string {
	return fmt.Sprintf(constants.UpdatePSHeader, repoPath) +
		fmt.Sprintf(constants.UpdatePSDeployDetect, repoPath) +
		constants.UpdatePSVersionBefore +
		fmt.Sprintf(constants.UpdatePSRunUpdate, runPS1) +
		constants.UpdatePSSync +
		constants.UpdatePSVersionAfter +
		fmt.Sprintf(constants.UpdatePSVerify, repoPath, repoPath) +
		constants.UpdatePSPostActions
}

// runUpdateScript executes the PowerShell script with output piped to terminal.
func runUpdateScript(scriptPath string) {
	cmd := exec.Command(constants.PSBin, constants.PSExecPolicy, constants.PSBypass,
		constants.PSNoProfile, constants.PSNoLogo, constants.PSFile, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	logScriptResult(err)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateFailed, err)
		os.Exit(1)
	}
}

// logScriptResult logs the update script exit status if verbose is active.
func logScriptResult(err error) {
	log := verbose.Get()
	if log != nil {
		log.Log(constants.UpdateScriptLogExit, err)
	}
}
