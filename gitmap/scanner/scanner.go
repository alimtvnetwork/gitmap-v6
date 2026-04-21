// Package scanner walks directories and detects Git repositories.
//
// The walker uses a small bounded worker pool so independent subtrees are
// crawled in parallel. On large folder trees this is I/O bound and yields
// a meaningful speedup; on small trees the pool collapses to effectively
// sequential work because the dispatch loop short-circuits when only one
// directory is in flight.
//
// Concurrency contract:
//   - Bounded by ScanWorkers (default = runtime.NumCPU(), capped by
//     scanWorkersMax to avoid pathological fd exhaustion on huge trees).
//   - Symlinks are NOT followed (consistent with the previous serial
//     implementation; see spec/01-app/03-scanner.md).
//   - When a `.git` directory is found the parent is recorded as a repo
//     and the subtree is NOT descended further (same rule as before).
//   - The first I/O error from any worker wins and is returned; remaining
//     workers drain and exit. Partial results discovered before the error
//     are still returned so callers can render what was found.
package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// scanWorkersMax caps the worker pool regardless of CPU count. Filesystem
// scans are I/O bound but each open dir consumes a file descriptor; 16 is
// well below the default ulimit on every supported platform.
const scanWorkersMax = 16

// MaxScanWorkers exposes the upper bound for callers (e.g. CLI flag
// validators) that want to clamp user-provided values into the supported
// range.
const MaxScanWorkers = scanWorkersMax

// RepoInfo holds raw data extracted from a discovered Git repo.
type RepoInfo struct {
	AbsolutePath string
	RelativePath string
}

// ScanDir walks root recursively and returns all Git repo paths found.
// Subtrees are crawled by a bounded worker pool sized via
// defaultWorkerCount(); result order is not guaranteed (callers that
// depend on lexical order must sort).
func ScanDir(root string, excludeDirs []string) ([]RepoInfo, error) {
	return ScanDirWithWorkers(root, excludeDirs, 0)
}

// ScanDirWithWorkers walks root using exactly `workers` goroutines.
// A value of 0 (or any negative number) selects the platform default
// from defaultWorkerCount(). Values larger than MaxScanWorkers are
// clamped down to keep the pool under the per-process fd budget.
func ScanDirWithWorkers(root string, excludeDirs []string, workers int) ([]RepoInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return walkParallel(absRoot, buildExcludeSet(excludeDirs), resolveWorkerCount(workers))
}

// resolveWorkerCount normalizes a caller-supplied worker count: 0 / <0
// means "auto", and any positive value is clamped to [1, MaxScanWorkers].
func resolveWorkerCount(requested int) int {
	if requested <= 0 {
		return defaultWorkerCount()
	}
	if requested > scanWorkersMax {
		return scanWorkersMax
	}

	return requested
}

// defaultWorkerCount picks a sensible pool size for the host CPU.
func defaultWorkerCount() int {
	n := runtime.NumCPU()
	if n < 1 {
		return 1
	}
	if n > scanWorkersMax {
		return scanWorkersMax
	}

	return n
}

// buildExcludeSet converts a slice to a set for O(1) lookups.
func buildExcludeSet(dirs []string) map[string]bool {
	set := make(map[string]bool, len(dirs))
	for _, d := range dirs {
		set[d] = true
	}

	return set
}

// scanState bundles the shared mutable state passed to every worker. It
// keeps the worker function tiny (well under the per-func line limit) and
// makes the synchronization rules obvious in one place.
type scanState struct {
	root    string
	exclude map[string]bool

	queue chan string  // pending directories to walk
	wg    sync.WaitGroup // tracks outstanding queued items, NOT workers

	mu     sync.Mutex
	repos  []RepoInfo
	firstErr error
}

// walkParallel runs a fixed-size worker pool that consumes directories
// from an unbounded-capacity FIFO and enqueues child directories back.
// The queue is closed when wg drops to zero — i.e. every dispatched
// directory has been fully processed and produced no new work.
func walkParallel(root string, exclude map[string]bool, workers int) ([]RepoInfo, error) {
	st := &scanState{
		root:    root,
		exclude: exclude,
		// Buffer sized generously so workers rarely block on enqueue.
		// A bounded buffer is fine — if it fills, workers backpressure
		// each other, which is acceptable; deadlock is impossible
		// because every send is paired with a wg.Add and the closer
		// only fires after wg.Done across all sends.
		queue: make(chan string, 1024),
	}

	st.wg.Add(1)
	st.queue <- root

	var workerWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for dir := range st.queue {
				st.processDir(dir)
				st.wg.Done()
			}
		}()
	}

	// Closer goroutine: once every queued dir has been processed (and
	// thus had a chance to enqueue its children), close the queue so
	// workers exit their range loop.
	go func() {
		st.wg.Wait()
		close(st.queue)
	}()

	workerWG.Wait()

	st.mu.Lock()
	defer st.mu.Unlock()

	return st.repos, st.firstErr
}

// processDir reads one directory and dispatches its child directories
// back onto the queue. Errors short-circuit further enqueues for THIS
// dir but do not stop other workers — the first error is captured and
// returned at the end.
func (st *scanState) processDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		st.recordErr(err)

		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		st.handleSubdir(dir, entry)
	}
}

// handleSubdir applies the exclude filter and the repo-detection rule,
// then either records the parent as a repo or enqueues the subdir for
// further walking.
func (st *scanState) handleSubdir(parent string, entry os.DirEntry) {
	name := entry.Name()
	if st.exclude[name] {
		return
	}
	if name == constants.ExtGit {
		st.recordRepo(parent)

		return
	}
	st.enqueue(filepath.Join(parent, name))
}

// enqueue dispatches a directory for processing.
func (st *scanState) enqueue(path string) {
	st.wg.Add(1)
	st.queue <- path
}

// recordRepo appends a discovered repo (parent of the .git dir) under
// the shared mutex. Repo recording is the only mutex contention point.
func (st *scanState) recordRepo(repoPath string) {
	rel, err := filepath.Rel(st.root, repoPath)
	if err != nil {
		st.recordErr(err)

		return
	}
	st.mu.Lock()
	st.repos = append(st.repos, RepoInfo{
		AbsolutePath: repoPath,
		RelativePath: rel,
	})
	st.mu.Unlock()
}

// recordErr stores the FIRST error to occur. Later errors are dropped to
// keep the public signature single-error and avoid a noisy multi-error.
func (st *scanState) recordErr(err error) {
	st.mu.Lock()
	if st.firstErr == nil {
		st.firstErr = err
	}
	st.mu.Unlock()
}
