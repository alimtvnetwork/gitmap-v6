// Package gitutil extracts Git metadata by running git commands.
package gitutil

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/user/gitmap/constants"
)

// RepoStatus holds the live state of a Git repository.
type RepoStatus struct {
	Branch       string
	Dirty        bool
	Untracked    int
	Modified     int
	Staged       int
	Ahead        int
	Behind       int
	StashCount   int
	Unreachable  bool
}

// RemoteURL returns the origin remote URL for a repo at the given path.
func RemoteURL(repoPath string) (string, error) {
	out, err := runGit(repoPath,
		constants.GitConfigCmd, constants.GitGetFlag, constants.GitRemoteOrigin)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// CurrentBranch returns the current branch name for a repo.
func CurrentBranch(repoPath string) (string, error) {
	out, err := runGit(repoPath,
		constants.GitRevParse, constants.GitAbbrevRef, constants.GitHEAD)
	if err != nil {
		return constants.DefaultBranch, err
	}

	return strings.TrimSpace(out), nil
}

// Status returns the full live status of a repository.
// If the path does not exist or is not a git repo, Unreachable is set.
func Status(repoPath string) RepoStatus {
	rs := RepoStatus{}

	if _, err := os.Stat(repoPath); err != nil {
		rs.Unreachable = true
		return rs
	}

	branch, err := CurrentBranch(repoPath)
	if err != nil {
		rs.Unreachable = true
		return rs
	}

	rs.Branch = branch
	rs.Dirty, rs.Untracked, rs.Modified, rs.Staged = parsePortcelainStatus(repoPath)
	rs.Ahead, rs.Behind = parseAheadBehind(repoPath)
	rs.StashCount = countStashes(repoPath)

	return rs
}

// parsePortcelainStatus runs git status --porcelain and counts file states.
func parsePortcelainStatus(repoPath string) (dirty bool, untracked, modified, staged int) {
	out, err := runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return false, 0, 0, 0
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		x := line[0]
		y := line[1]
		if x == '?' && y == '?' {
			untracked++
		} else if x != ' ' && x != '?' {
			staged++
		}
		if y != ' ' && y != '?' {
			modified++
		}
	}
	dirty = (untracked + modified + staged) > 0

	return dirty, untracked, modified, staged
}

// parseAheadBehind extracts ahead/behind counts from rev-list.
func parseAheadBehind(repoPath string) (ahead, behind int) {
	out, err := runGit(repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) == 2 {
		ahead, _ = strconv.Atoi(parts[0])
		behind, _ = strconv.Atoi(parts[1])
	}

	return ahead, behind
}

// countStashes returns the number of stash entries.
func countStashes(repoPath string) int {
	out, err := runGit(repoPath, "stash", "list")
	if err != nil || len(strings.TrimSpace(out)) == 0 {
		return 0
	}

	return len(strings.Split(strings.TrimSpace(out), "\n"))
}

// runGit executes a git command in the given directory and returns stdout.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command(constants.GitBin, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
