// Package scanner walks directories and detects Git repositories.
package scanner

import (
	"os"
	"path/filepath"

	"github.com/user/gitmap/constants"
)

// RepoInfo holds raw data extracted from a discovered Git repo.
type RepoInfo struct {
	AbsolutePath string
	RelativePath string
}

// ScanDir walks root recursively and returns all Git repo paths found.
func ScanDir(root string, excludeDirs []string) ([]RepoInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return walkTree(absRoot, absRoot, buildExcludeSet(excludeDirs))
}

// buildExcludeSet converts a slice to a set for O(1) lookups.
func buildExcludeSet(dirs []string) map[string]bool {
	set := make(map[string]bool, len(dirs))
	for _, d := range dirs {
		set[d] = true
	}

	return set
}

// walkTree recursively walks the directory tree from current.
func walkTree(root, current string, exclude map[string]bool) ([]RepoInfo, error) {
	var repos []RepoInfo
	entries, err := os.ReadDir(current)
	if err != nil {
		return repos, err
	}

	return processEntries(root, current, entries, exclude, repos)
}

// processEntries iterates directory entries and collects repos.
func processEntries(
	root, current string,
	entries []os.DirEntry,
	exclude map[string]bool,
	repos []RepoInfo,
) ([]RepoInfo, error) {
	for _, entry := range entries {
		if entry.IsDir() {
			found, err := handleDir(root, current, entry, exclude)
			if err != nil {
				return repos, err
			}
			repos = append(repos, found...)
		}
	}

	return repos, nil
}

// handleDir decides whether a directory is a repo, excluded, or walkable.
func handleDir(
	root, current string,
	entry os.DirEntry,
	exclude map[string]bool,
) ([]RepoInfo, error) {
	name := entry.Name()
	fullPath := filepath.Join(current, name)

	if exclude[name] {
		return nil, nil
	}
	if name == constants.ExtGit {
		return foundRepo(root, current)
	}

	return walkTree(root, fullPath, exclude)
}

// foundRepo creates a RepoInfo for the current directory (parent of .git).
func foundRepo(root, repoPath string) ([]RepoInfo, error) {
	rel, err := filepath.Rel(root, repoPath)
	if err != nil {
		return nil, err
	}
	info := RepoInfo{
		AbsolutePath: repoPath,
		RelativePath: rel,
	}

	return []RepoInfo{info}, nil
}
