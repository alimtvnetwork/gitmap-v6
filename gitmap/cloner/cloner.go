// Package cloner re-clones repos from structured files.
package cloner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/formatter"
	"github.com/user/gitmap/model"
)

// CloneFromFile reads a source file and clones all repos under targetDir.
func CloneFromFile(sourcePath, targetDir string, safePull bool) (model.CloneSummary, error) {
	records, err := loadRecords(sourcePath)
	if err != nil {
		return model.CloneSummary{}, err
	}

	return cloneAll(records, targetDir, safePull, false), nil
}

// CloneFromFileQuiet reads a source file and clones with suppressed progress.
func CloneFromFileQuiet(sourcePath, targetDir string, safePull bool) (model.CloneSummary, error) {
	records, err := loadRecords(sourcePath)
	if err != nil {
		return model.CloneSummary{}, err
	}

	return cloneAll(records, targetDir, safePull, true), nil
}

// loadRecords detects file format and parses records.
func loadRecords(path string) ([]model.ScanRecord, error) {
	ext := strings.ToLower(filepath.Ext(path))
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseByExtension(ext, file)
}

// parseByExtension dispatches to the correct parser.
func parseByExtension(ext string, r io.Reader) ([]model.ScanRecord, error) {
	if ext == constants.ExtCSV {
		return formatter.ParseCSV(r)
	}
	if ext == constants.ExtJSON {
		return formatter.ParseJSON(r)
	}

	return parseTextFile(r)
}

// parseTextFile reads one git clone command per line.
func parseTextFile(r io.Reader) ([]model.ScanRecord, error) {
	var records []model.ScanRecord
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if len(line) > 0 {
			records = append(records, parseCloneLine(line))
		}
	}

	return records, sc.Err()
}

// parseCloneLine extracts url, branch, path from a git clone command.
func parseCloneLine(line string) model.ScanRecord {
	parts := strings.Fields(line)
	rec := model.ScanRecord{CloneInstruction: line}
	if len(parts) >= 5 {
		rec.Branch = parts[3]
		rec.HTTPSUrl = parts[4]
	}
	if len(parts) >= 6 {
		rec.RelativePath = parts[5]
	}

	return rec
}

// cloneAll iterates records and clones each one with progress tracking.
func cloneAll(records []model.ScanRecord, targetDir string, safePull, quiet bool) model.CloneSummary {
	if !safePull && hasExistingRepos(records, targetDir) {
		safePull = true
		fmt.Print(constants.MsgAutoSafePull)
	}

	progress := NewProgress(len(records), quiet)
	summary := model.CloneSummary{}

	for _, rec := range records {
		progress.Begin(repoDisplayName(rec))
		result := cloneOrPullOne(rec, targetDir, safePull)
		trackResult(progress, result, rec, targetDir, safePull)
		summary = updateSummary(summary, result)
	}

	progress.PrintSummary()

	return summary
}

// repoDisplayName returns a display name for progress output.
func repoDisplayName(rec model.ScanRecord) string {
	if len(rec.RepoName) > 0 {
		return rec.RepoName
	}

	return rec.RelativePath
}

// trackResult updates progress based on clone/pull outcome.
func trackResult(p *Progress, result model.CloneResult, rec model.ScanRecord, targetDir string, safePull bool) {
	if result.Success {
		pulled := safePull && isGitRepo(filepath.Join(targetDir, rec.RelativePath))
		p.Done(result, pulled)

		return
	}

	p.Fail(result)
}

// hasExistingRepos checks if any target repo directories already exist.
func hasExistingRepos(records []model.ScanRecord, targetDir string) bool {
	for _, rec := range records {
		dest := filepath.Join(targetDir, rec.RelativePath)
		if isGitRepo(dest) {
			return true
		}
	}

	return false
}

// cloneOne clones a single repository.
func cloneOne(rec model.ScanRecord, targetDir string) model.CloneResult {
	dest := filepath.Join(targetDir, rec.RelativePath)
	err := os.MkdirAll(filepath.Dir(dest), constants.DirPermission)
	if err != nil {
		return model.CloneResult{Record: rec, Success: false, Error: err.Error()}
	}

	return runClone(rec, dest)
}

// runClone executes the git clone command.
func runClone(rec model.ScanRecord, dest string) model.CloneResult {
	url := pickURL(rec)
	cmd := exec.Command(constants.GitBin, constants.GitClone,
		constants.GitBranchFlag, rec.Branch, url, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("%s: %s", err.Error(), string(out))

		return model.CloneResult{Record: rec, Success: false, Error: msg}
	}

	return model.CloneResult{Record: rec, Success: true}
}

// pickURL selects the best available URL from a record.
func pickURL(rec model.ScanRecord) string {
	if len(rec.HTTPSUrl) > 0 {
		return rec.HTTPSUrl
	}

	return rec.SSHUrl
}

// updateSummary increments counters and collects results.
func updateSummary(s model.CloneSummary, r model.CloneResult) model.CloneSummary {
	if r.Success {
		s.Succeeded++
		s.Cloned = append(s.Cloned, r)

		return s
	}
	s.Failed++
	s.Errors = append(s.Errors, r)

	return s
}
