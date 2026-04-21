package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/verbose"
)

// runUpdate handles the "update" subcommand.
// It creates a handoff copy and runs a hidden worker command from that copy.
func runUpdate() {
	requireOnline()
	repoPath := resolveRepoPath()
	report := resolveReportErrors()
	report.announce()

	selfPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateExecFind, err)
		os.Exit(1)
	}

	copyPath := createHandoffCopy(selfPath)
	fmt.Printf(constants.MsgUpdateActive, selfPath, copyPath)
	launchHandoff(copyPath, repoPath, report)
}

// resolveRepoPath returns the repo path from --repo-path flag or embedded constant.
// If neither is available, it attempts to delegate to gitmap-updater.
func resolveRepoPath() string {
	for _, repoPath := range []string{
		resolveRepoPathFromFlag(),
		resolveRepoPathFromEmbedded(),
		resolveRepoPathFromDB(),
	} {
		if len(repoPath) == 0 {
			continue
		}

		saveRepoPathToDB(repoPath)

		return repoPath
	}

	if prompted := promptRepoPath(); len(prompted) > 0 {
		saveRepoPathToDB(prompted)

		return prompted
	}

	// Try to fall back to gitmap-updater for release-based update
	if tryUpdaterFallback() {
		os.Exit(0)
	}

	fmt.Fprint(os.Stderr, constants.ErrNoRepoPath)
	os.Exit(1)

	return ""
}

// tryUpdaterFallback looks for gitmap-updater on PATH and launches it.
func tryUpdaterFallback() bool {
	updaterPath, err := exec.LookPath(constants.UpdaterBin)
	if err != nil {
		return false
	}

	fmt.Printf(constants.MsgUpdaterFallback, updaterPath)
	cmd := exec.Command(updaterPath, "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		return false
	}

	return true
}

// createHandoffCopy creates a temporary copy of the binary for handoff.
func createHandoffCopy(selfPath string) string {
	nameFmt := constants.UpdateCopyFmtExe
	if runtime.GOOS != "windows" {
		nameFmt = constants.UpdateCopyFmtUnix
	}

	name := fmt.Sprintf(nameFmt, os.Getpid())
	copyPath := filepath.Join(filepath.Dir(selfPath), name)

	if copyFile(selfPath, copyPath) == nil {
		makeExecutable(copyPath)

		return copyPath
	}

	fallbackPath := filepath.Join(os.TempDir(), name)
	if err := copyFile(selfPath, fallbackPath); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrUpdateCopyFail, err)
		os.Exit(1)
	}

	makeExecutable(fallbackPath)

	return fallbackPath
}

// makeExecutable sets executable permission on Unix systems.
func makeExecutable(path string) {
	if runtime.GOOS == "windows" {
		return
	}

	if err := os.Chmod(path, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not make %s executable: %v\n", path, err)
	}
}

// launchHandoff runs the handoff binary with update-runner command.
func launchHandoff(copyPath, repoPath string, report reportErrorsConfig) {
	copyArgs := []string{constants.CmdUpdateRunner}
	if hasFlag(constants.FlagVerbose) {
		copyArgs = append(copyArgs, constants.FlagVerbose)
	}

	copyArgs = append(copyArgs, constants.FlagRepoPath, repoPath)
	copyArgs = report.applyToHandoffArgs(copyArgs)

	cmd := exec.Command(copyPath, copyArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = report.applyToEnv(os.Environ())
	if err := cmd.Run(); err != nil {
		handleHandoffError(err)
	}
}

// handleHandoffError exits with the handoff process exit code if available.
func handleHandoffError(err error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}

	fmt.Fprintf(os.Stderr, constants.ErrUpdateFailed, err)
	os.Exit(1)
}

// runUpdateRunner is a hidden command that performs the real update work.
// After the binary update completes, it triggers a best-effort schema
// migration so the next CLI invocation never has to repair the database.
//
// At the end, scheduleHandoffSelfDelete arranges for the running handoff
// copy (gitmap-update-<pid>.exe) to remove itself after exit. This is
// required on Windows because the binary is locked while running and
// `gitmap update-cleanup` cannot delete it from inside the same process
// tree. See spec/02-app-issues/01-update-file-lock.md.
func runUpdateRunner() {
	repoPath := resolveRepoPath()
	report := resolveReportErrors()

	initRunnerVerbose()
	fmt.Printf(constants.MsgUpdateStarting)
	fmt.Printf(constants.MsgUpdateRepoPath, repoPath)
	executeUpdate(repoPath, report)
	runPostUpdateMigrate()
	report.summarize()
	scheduleHandoffSelfDelete()
}

// scheduleHandoffSelfDelete schedules deletion of the running handoff
// binary after this process exits. On Windows we spawn a detached
// cmd.exe that pings briefly (so our process can exit and release the
// file lock) and then `del`s the file. On Unix we just unlink in place.
//
// Best-effort: silent on failure. Only runs when the active executable
// matches the handoff naming pattern, so a normally-deployed gitmap.exe
// is never touched even if invoked with `update-runner` manually.
func scheduleHandoffSelfDelete() {
	self, err := os.Executable()
	if err != nil {
		return
	}

	base := filepath.Base(self)
	if !isUpdateHandoffName(base) {
		return
	}

	if runtime.GOOS != constants.OSWindows {
		_ = os.Remove(self)

		return
	}

	cmd := exec.Command("cmd.exe", "/C",
		"ping", "127.0.0.1", "-n", "2", ">nul", "&", "del", "/F", "/Q", self)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()
}

// isUpdateHandoffName reports whether base looks like a handoff copy
// produced by createHandoffCopy (e.g. "gitmap-update-22112.exe").
func isUpdateHandoffName(base string) bool {
	const prefix = "gitmap-update-"

	return len(base) > len(prefix) && base[:len(prefix)] == prefix
}

// getFlagValue returns the value following a flag like --repo-path <value>.
func getFlagValue(name string) string {
	args := os.Args[2:]
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

// initRunnerVerbose initializes verbose logging if --verbose flag is present.
func initRunnerVerbose() {
	if hasFlag(constants.FlagVerbose) {
		log, err := verbose.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.WarnVerboseLogFailed, err)
		} else {
			defer log.Close()
			log.Log(constants.UpdateRunnerLogStart, constants.RepoPath)
		}
	}
}

// hasFlag checks if a flag is present in os.Args[2:].
func hasFlag(name string) bool {
	for _, arg := range os.Args[2:] {
		if arg == name {
			return true
		}
	}

	return false
}

// copyFile copies src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	return err
}
