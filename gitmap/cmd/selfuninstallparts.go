package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/setup"
)

// selfDeployDir returns the directory the live binary is installed in.
// Falls back to "" if it cannot be resolved.
func selfDeployDir() string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(self)
	if err != nil {
		resolved = self
	}

	return filepath.Dir(resolved)
}

// selfDataDir returns the .gitmap data directory anchored to the binary.
func selfDataDir() string {
	deploy := selfDeployDir()
	if len(deploy) == 0 {
		return ""
	}

	return filepath.Join(deploy, constants.DefaultOutputFolder)
}

// removeDeployArtifacts deletes the gitmap binary plus any sibling
// artefacts (handoff temp copies, .old backups, completion files).
func removeDeployArtifacts(dir string) {
	if len(dir) == 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgSelfUninstallSkipBin, err)

		return
	}
	for _, e := range entries {
		if !isGitmapArtifact(e.Name()) {
			continue
		}
		full := filepath.Join(dir, e.Name())
		removePathBestEffort(full)
	}
}

// isGitmapArtifact reports whether a filename belongs to the gitmap
// install (binary, handoff copies, completion outputs, .old backups).
func isGitmapArtifact(name string) bool {
	lower := strings.ToLower(name)
	if lower == "gitmap" || lower == "gitmap.exe" {
		return true
	}
	if strings.HasPrefix(lower, "gitmap-handoff-") {
		return true
	}
	if strings.HasSuffix(lower, ".old") && strings.HasPrefix(lower, "gitmap") {
		return true
	}
	if strings.HasPrefix(lower, "gitmap-completion") {
		return true
	}

	return false
}

// removeCompletionFiles wipes generated bash/zsh/fish completion files
// that gitmap may have written under the deploy dir.
func removeCompletionFiles(dir string) {
	if len(dir) == 0 {
		return
	}
	candidates := []string{"gitmap-completion.bash", "gitmap-completion.zsh", "gitmap-completion.fish"}
	for _, c := range candidates {
		removePathBestEffort(filepath.Join(dir, c))
	}
}

// removePathBestEffort deletes a file or directory and reports the
// outcome. Missing paths are silently ignored.
func removePathBestEffort(path string) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfUninstallRemove, path, err)

		return
	}
	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfUninstallRemove, path, err)

		return
	}
	if info.IsDir() {
		fmt.Printf(constants.MsgSelfUninstallRemovedDir, path)

		return
	}
	fmt.Printf(constants.MsgSelfUninstallRemovedBin, path)
}

// removeProfileSnippet strips the gitmap shell-wrapper marker block
// from the user's shell profile, leaving the rest of the file intact.
func removeProfileSnippet(profile string) {
	if len(profile) == 0 {
		return
	}
	data, err := os.ReadFile(profile)
	if os.IsNotExist(err) {
		fmt.Printf(constants.MsgSelfUninstallSnippetMiss, profile)

		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfUninstallSnippetRead, profile, err)

		return
	}
	stripped, removed := stripMarkerBlock(string(data))
	if !removed {
		fmt.Printf(constants.MsgSelfUninstallSnippetMiss, profile)

		return
	}
	if writeErr := os.WriteFile(profile, []byte(stripped), 0o644); writeErr != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfUninstallSnippetWrite, profile, writeErr)

		return
	}
	fmt.Printf(constants.MsgSelfUninstallSnippetGone, profile)
}

// stripMarkerBlock removes any line range delimited by the gitmap
// shell-wrapper marker open/close lines, regardless of manager string.
func stripMarkerBlock(content string) (string, bool) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var out strings.Builder
	skip := false
	removed := false
	for scanner.Scan() {
		line := scanner.Text()
		if !skip && isMarkerOpen(line) {
			skip = true
			removed = true

			continue
		}
		if skip && line == setup.MarkerClose() {
			skip = false

			continue
		}
		if !skip {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}
	if !strings.HasSuffix(content, "\n") {
		return strings.TrimRight(out.String(), "\n"), removed
	}

	return out.String(), removed
}

// isMarkerOpen matches the marker open line for any manager string.
func isMarkerOpen(line string) bool {
	return strings.HasPrefix(line, "# gitmap shell wrapper v") &&
		strings.Contains(line, " - managed by ")
}

// defaultProfileForOS picks the conventional rc file for the current OS.
func defaultProfileForOS() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if isWindows() {
		return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	return filepath.Join(home, ".bashrc")
}
