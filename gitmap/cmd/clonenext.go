package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/gitmap/clonenext"
	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/desktop"
	"github.com/user/gitmap/gitutil"
	"github.com/user/gitmap/lockcheck"
	"github.com/user/gitmap/model"
	"github.com/user/gitmap/verbose"
)

// runCloneNext handles the "clone-next" subcommand.
func runCloneNext(args []string) {
	checkHelp("clone-next", args)
	versionArg, deleteFlag, keepFlag, noDesktop, createRemote, sshKeyName, verboseMode := parseCloneNextFlags(args)
	if len(versionArg) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrCloneNextUsage)
		os.Exit(1)
	}

	if verboseMode {
		log, err := verbose.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.WarnVerboseLogFailed, err)
		} else {
			defer log.Close()
		}
	}

	requireOnline()
	applySSHKey(sshKeyName)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextCwd, err)
		os.Exit(1)
	}

	remoteURL, err := gitutil.RemoteURL(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextNoRemote, err)
		os.Exit(1)
	}

	currentFolder := filepath.Base(cwd)
	parentDir := filepath.Dir(cwd)

	// Strip .git suffix from remote URL for repo name extraction.
	repoName := extractRepoName(remoteURL)

	parsed := clonenext.ParseRepoName(repoName)
	targetVersion, err := clonenext.ResolveTarget(parsed, versionArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextBadVersion, err)
		os.Exit(1)
	}

	targetName := clonenext.TargetRepoName(parsed.BaseName, targetVersion)
	targetURL := clonenext.ReplaceRepoInURL(remoteURL, repoName, targetName)

	// Flatten by default: clone into base name folder (no version suffix).
	flattenedFolder := parsed.BaseName
	targetPath := filepath.Join(parentDir, flattenedFolder)

	// If the flattened folder already exists, try to remove it for a fresh clone.
	// On Windows, the current shell's working directory is locked and cannot be
	// removed by this process. In that case, fall back to a versioned folder name
	// (e.g. scripts-fixer-v2) and warn — never abort the whole flow.
	if _, statErr := os.Stat(targetPath); statErr == nil {
		fmt.Printf(constants.MsgFlattenRemoving, flattenedFolder)
		if removeErr := os.RemoveAll(targetPath); removeErr != nil {
			fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, flattenedFolder, removeErr)
			fallbackFolder := targetName
			fallbackPath := filepath.Join(parentDir, fallbackFolder)
			fmt.Printf(constants.MsgFlattenFallback, fallbackFolder)
			fmt.Printf(constants.MsgFlattenLockedHint, flattenedFolder)
			// If the versioned fallback also exists, attempt to remove it; if that
			// fails too, warn but continue — git clone will surface a clear error.
			if _, fbStat := os.Stat(fallbackPath); fbStat == nil {
				if fbErr := os.RemoveAll(fallbackPath); fbErr != nil {
					fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, fallbackFolder, fbErr)
				}
			}
			flattenedFolder = fallbackFolder
			targetPath = fallbackPath
		}
	}

	// Optionally check and create the target GitHub repo when --create-remote is set.
	if createRemote {
		owner, _, parseErr := clonenext.ParseOwnerRepo(remoteURL)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, constants.ErrCloneNextRemoteParse, parseErr)
			os.Exit(1)
		}

		exists, checkErr := clonenext.RepoExists(owner, targetName)
		if checkErr != nil {
			fmt.Fprintf(os.Stderr, constants.ErrCloneNextRepoCheck, checkErr)
			os.Exit(1)
		}

		if !exists {
			fmt.Printf(constants.MsgCloneNextCreating, targetName)
			createErr := clonenext.CreateRepo(owner, targetName, true)
			if createErr != nil {
				fmt.Fprintf(os.Stderr, constants.ErrCloneNextRepoCreate, targetName, createErr)
				os.Exit(1)
			}
			fmt.Printf(constants.MsgCloneNextCreated, targetName)
		}
	}

	fmt.Printf(constants.MsgFlattenCloning, targetName, flattenedFolder)
	cloneResult := runGitClone(targetURL, targetPath)
	if !cloneResult {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextFailed, targetName)
		os.Exit(1)
	}
	fmt.Printf(constants.MsgFlattenDone, targetName, flattenedFolder)

	// Record version history in DB.
	recordVersionHistory(targetPath, parsed.CurrentVersion, targetVersion, flattenedFolder)

	if !noDesktop {
		registerCloneNextDesktop(targetName, targetPath)
	}

	// Handle removal of the old versioned folder (only if different from flattened path).
	if currentFolder != flattenedFolder {
		handleCloneNextRemoval(currentFolder, cwd, targetPath, deleteFlag, keepFlag)
	}

	// Set GITMAP_SHELL_HANDOFF for the shell wrapper to cd into the new folder.
	os.Setenv("GITMAP_SHELL_HANDOFF", targetPath)

	// Open in VS Code if available.
	openInVSCode(targetPath)
}

// extractRepoName extracts the repository name from a remote URL.
func extractRepoName(remoteURL string) string {
	name := remoteURL
	// Remove trailing .git
	name = strings.TrimSuffix(name, ".git")
	// Get last path segment
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}

	return name
}

// runGitClone executes git clone and returns success status.
func runGitClone(url, dest string) bool {
	cmd := exec.Command(constants.GitBin, constants.GitClone, url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() == nil
}

// registerCloneNextDesktop registers the cloned repo with GitHub Desktop.
func registerCloneNextDesktop(name, absPath string) {
	records := []model.ScanRecord{{
		RepoName:     name,
		AbsolutePath: absPath,
	}}
	result := desktop.AddRepos(records)
	if result.Added > 0 {
		fmt.Printf(constants.MsgCloneNextDesktop, name)
	}
}

// handleCloneNextRemoval manages removal of the current version folder.
// It changes to the parent directory first to release file locks on Windows.
func handleCloneNextRemoval(folderName, fullPath, targetPath string, deleteFlag, keepFlag bool) {
	if keepFlag {
		return
	}

	// Move out of the folder before attempting removal to avoid Windows file locks.
	parentDir := filepath.Dir(fullPath)
	if chErr := os.Chdir(parentDir); chErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not cd to %s: %v\n", parentDir, chErr)
	}

	removed := false
	var shouldRemove bool

	if deleteFlag {
		shouldRemove = true
	} else {
		// Prompt
		fmt.Printf(constants.MsgCloneNextRemovePrompt, folderName)
		var answer string
		_, _ = fmt.Scanln(&answer)
		shouldRemove = strings.ToLower(strings.TrimSpace(answer)) == "y"
	}

	if shouldRemove {
		removed = removeFolderWithLockCheck(folderName, fullPath)
	}

	// After removing the old folder, move into the newly cloned directory.
	if removed {
		if chErr := os.Chdir(targetPath); chErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not cd to %s: %v\n", targetPath, chErr)
		} else {
			fmt.Printf(constants.MsgCloneNextMovedTo, filepath.Base(targetPath))
		}
	}
}

// removeFolderWithLockCheck attempts to remove a directory, and if it fails,
// scans for locking processes and offers to terminate them before retrying.
// All removal attempts are tracked as pending tasks in the database.
func removeFolderWithLockCheck(name, path string) bool {
	// Record the delete intent as a pending task before any OS operation.
	taskID, db := createPendingTask(constants.TaskTypeDelete, path, "", constants.CmdCloneNext, "")
	if db != nil {
		defer db.Close()
	}

	// First attempt.
	err := os.RemoveAll(path)
	if err == nil {
		fmt.Printf(constants.MsgCloneNextRemoved, name)
		completePendingTask(db, taskID)

		return true
	}

	// Removal failed — scan for locking processes.
	fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, name, err)
	fmt.Printf(constants.MsgLockCheckScanning, name)

	procs, scanErr := lockcheck.FindLockingProcesses(path)
	if scanErr != nil {
		fmt.Fprintf(os.Stderr, constants.WarnLockCheckScanFailed, scanErr)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonLockScanFailed, scanErr))

		return false
	}

	if len(procs) == 0 {
		fmt.Print(constants.MsgLockCheckNoneFound)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonNoLockingProcs, err))

		return false
	}

	// Show locking processes and prompt to kill.
	fmt.Printf(constants.MsgLockCheckFound, lockcheck.FormatProcessList(procs))
	fmt.Print(constants.MsgLockCheckKillPrompt)

	var answer string
	_, _ = fmt.Scanln(&answer)
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		failPendingTask(db, taskID, constants.ReasonUserDeclined)

		return false
	}

	// Terminate each process.
	for _, p := range procs {
		fmt.Printf(constants.MsgLockCheckKilling, p.Name, p.PID)
		killErr := lockcheck.KillProcess(p.PID)
		if killErr != nil {
			fmt.Fprintf(os.Stderr, constants.WarnLockCheckKillFailed, p.Name, p.PID, killErr)
		} else {
			fmt.Printf(constants.MsgLockCheckKilled, p.Name)
		}
	}

	// Brief pause to let OS release handles.
	time.Sleep(500 * time.Millisecond)

	// Retry removal.
	fmt.Print(constants.MsgLockCheckRetrying)
	retryErr := os.RemoveAll(path)
	if retryErr != nil {
		fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, name, retryErr)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonRetryFailed, retryErr))

		return false
	}

	fmt.Printf(constants.MsgCloneNextRemoved, name)
	completePendingTask(db, taskID)

	return true
}
