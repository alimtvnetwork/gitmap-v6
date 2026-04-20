// Package release handles version parsing, release workflows,
// GitHub integration, and release metadata management.
package release

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/user/gitmap/constants"
)

// ChangelogEntry represents one version section in CHANGELOG.md.
type ChangelogEntry struct {
	Version string
	Notes   []string
}

// ReadChangelog reads concise changelog entries from CHANGELOG.md.
func ReadChangelog() ([]ChangelogEntry, error) {
	file, err := os.Open(constants.ChangelogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var entries []ChangelogEntry
	current := ChangelogEntry{}
	inSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") {
			if inSection {
				entries = append(entries, current)
			}
			version := parseVersionHeader(line)
			if len(version) == 0 {
				inSection = false
				continue
			}
			current = ChangelogEntry{Version: version, Notes: []string{}}
			inSection = true
			continue
		}

		if inSection {
			if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
				note := strings.TrimSpace(line[2:])
				if len(note) > 0 {
					current.Notes = append(current.Notes, note)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if inSection {
		entries = append(entries, current)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no version sections found in %s", constants.ChangelogFile)
	}

	return entries, nil
}

// FindChangelogEntry returns a changelog entry by version.
func FindChangelogEntry(entries []ChangelogEntry, version string) (ChangelogEntry, bool) {
	target := NormalizeVersion(version)
	for _, entry := range entries {
		if NormalizeVersion(entry.Version) == target {
			return entry, true
		}
	}

	return ChangelogEntry{}, false
}

// NormalizeVersion normalizes a changelog version string to v-prefixed form.
func NormalizeVersion(version string) string {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "gitmap")
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		return ""
	}
	if strings.HasPrefix(v, "v") {
		return v
	}

	return "v" + v
}

// parseVersionHeader extracts the version token from a markdown heading.
func parseVersionHeader(header string) string {
	raw := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	if len(raw) == 0 {
		return ""
	}

	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ""
	}

	version := strings.Trim(parts[0], "[]")
	if len(version) == 0 {
		return ""
	}
	if strings.HasPrefix(version, "v") {
		return version
	}

	return "v" + version
}
