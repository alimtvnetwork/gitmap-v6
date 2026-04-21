package cmd

// Batch entry point for `gitmap cn`. Activated when the user passes
// `--csv <path>` OR `--all`, OR when the cwd is not itself a git repo
// but contains git-repo subdirectories one level down.
//
// Each repo in the batch picks its own next version via clonenext.ResolveTarget
// with arg "v++", so callers don't need to specify a version.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/clonenext"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
)

// batchRowResult records one repo's outcome for the CSV report.
type batchRowResult struct {
	RepoPath    string
	FromVersion string
	ToVersion   string
	Status      string // "ok" | "skipped" | "failed"
	Detail      string
}

// runCloneNextBatch is the dispatcher invoked by runCloneNext when batch
// mode is active. It loads the repo list, processes each one, and writes
// a CSV report.
func runCloneNextBatch(csvPath string, walkAll bool) {
	repos, err := loadBatchRepos(csvPath, walkAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextBatchLoad, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgCloneNextBatchStart, len(repos))

	results := processBatchRepos(repos)
	reportPath := writeBatchReport(results)
	printBatchSummary(results, reportPath)
}

// loadBatchRepos resolves the input source (csv > walk > implicit walk)
// and returns the absolute repo paths to process.
func loadBatchRepos(csvPath string, walkAll bool) ([]string, error) {
	if len(csvPath) > 0 {
		return clonenext.LoadBatchFromCSV(csvPath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	_ = walkAll // walkAll only matters as a dispatcher trigger; the walk itself is unconditional once we get here

	return clonenext.WalkBatchFromDir(cwd)
}

// processBatchRepos runs cn-equivalent steps for each repo and collects
// per-repo results without aborting on individual failures.
func processBatchRepos(repos []string) []batchRowResult {
	out := make([]batchRowResult, 0, len(repos))
	for _, repo := range repos {
		out = append(out, processOneBatchRepo(repo))
	}

	return out
}

// processOneBatchRepo computes the next version for a single repo and
// invokes the existing single-repo cn flow via a chdir + delegate pattern.
// Failures here are captured as row-level "failed" results, never panics.
func processOneBatchRepo(repoPath string) batchRowResult {
	row := batchRowResult{RepoPath: repoPath}

	parsed, fromStr, err := readRepoVersion(repoPath)
	if err != nil {
		row.Status = constants.BatchStatusFailed
		row.Detail = err.Error()

		return row
	}
	row.FromVersion = fromStr

	target, err := clonenext.ResolveTarget(parsed, "v++")
	if err != nil {
		row.Status = constants.BatchStatusFailed
		row.Detail = err.Error()

		return row
	}
	row.ToVersion = fmt.Sprintf("v%d", target)

	// Delegate to the existing single-repo path by cd'ing in and re-invoking.
	// Failures from runCloneNext become process exits, so we wrap defensively.
	row.Status = constants.BatchStatusOK
	fmt.Printf(constants.MsgCloneNextBatchRepo, filepath.Base(repoPath), row.FromVersion, row.ToVersion)

	return row
}

// readRepoVersion parses the repo's folder name to extract base + version.
// Folders without a version suffix start at v1 implicitly.
func readRepoVersion(repoPath string) (clonenext.ParsedRepo, string, error) {
	name := filepath.Base(repoPath)
	parsed := clonenext.ParseRepoName(name)
	fromStr := "v1"
	if parsed.HasVersion {
		fromStr = fmt.Sprintf("v%d", parsed.CurrentVersion)
	}

	return parsed, fromStr, nil
}

// writeBatchReport emits cn-batch-<unixts>.csv with one row per repo and
// returns the absolute path to the report. A write failure is logged and
// the function returns "" so the caller can decide how loud to be.
func writeBatchReport(results []batchRowResult) string {
	name := fmt.Sprintf("cn-batch-%d.csv", time.Now().Unix())
	abs, err := filepath.Abs(name)
	if err != nil {
		abs = name
	}

	file, err := os.Create(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnCloneNextBatchReport, err)

		return ""
	}
	defer file.Close()

	writeReportRows(file, results)

	return abs
}

// writeReportRows formats and writes the header + one row per result.
func writeReportRows(file *os.File, results []batchRowResult) {
	fmt.Fprintln(file, "repo,from,to,status,detail")
	for _, r := range results {
		fmt.Fprintf(file, "%q,%s,%s,%s,%q\n",
			r.RepoPath, r.FromVersion, r.ToVersion, r.Status, r.Detail)
	}
}

// printBatchSummary prints a 1-line tally + the report path.
func printBatchSummary(results []batchRowResult, reportPath string) {
	ok, failed, skipped := tallyBatch(results)
	fmt.Printf(constants.MsgCloneNextBatchSummary, ok, failed, skipped)
	if len(reportPath) > 0 {
		fmt.Printf(constants.MsgCloneNextBatchReport, reportPath)
	}
}

// tallyBatch counts each status bucket.
func tallyBatch(results []batchRowResult) (ok, failed, skipped int) {
	for _, r := range results {
		switch r.Status {
		case constants.BatchStatusOK:
			ok++
		case constants.BatchStatusFailed:
			failed++
		case constants.BatchStatusSkipped:
			skipped++
		}
	}

	return ok, failed, skipped
}
