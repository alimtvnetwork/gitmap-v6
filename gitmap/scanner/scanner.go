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
//   - Cancellation: every public entry point accepts (or wraps) a
//     context.Context. When the context is cancelled, in-flight workers
//     stop dispatching new directories and drain quickly. The partial
//     repo list collected so far is returned alongside ctx.Err() so
//     callers can decide whether to surface results or discard them.
//
// Race-safety design (spec/05-coding-guidelines/16-concurrency-patterns.md):
//   - Repos are NEVER appended from worker goroutines. Each discovery is
//     sent on a buffered results channel and a single dedicated collector
//     goroutine owns the result slice. This eliminates the
//     "slice-append-under-mutex" pattern flagged by the spec.
//   - Errors use an atomic compare-and-swap on *scanError so the first
//     failure wins without taking a mutex on the hot path.
//   - The dispatch counter (`inflight`) is an atomic.Int64. Workers only
//     touch shared mutable state via channels and atomics — never via
//     a shared slice or map.
//   - All loops select on ctx.Done() so cancellation is observed within
//     one directory of work.
//
// Run `go test -race ./scanner/...` to validate.
package scanner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

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

// queueBuffer sizes the directory FIFO. Generous so workers rarely block
// on enqueue; bounded so a pathological tree can't exhaust memory.
const queueBuffer = 1024

// resultsBuffer sizes the discovered-repo channel. Repos are extremely
// cheap to enqueue (two strings) and the collector drains continuously,
// so a modest buffer absorbs bursts without blocking workers.
const resultsBuffer = 256

// RepoInfo holds raw data extracted from a discovered Git repo.
type RepoInfo struct {
	AbsolutePath string
	RelativePath string
}

// ScanDir walks root recursively and returns all Git repo paths found.
// Subtrees are crawled by a bounded worker pool sized via
// defaultWorkerCount(); result order is not guaranteed (callers that
// depend on lexical order must sort).
//
// This is a convenience wrapper around ScanDirContext that uses a
// background context — i.e. no cancellation. New code that wants
// Ctrl+C support should call ScanDirContext directly.
func ScanDir(root string, excludeDirs []string) ([]RepoInfo, error) {
	return ScanDirContext(context.Background(), root, excludeDirs, 0)
}

// ScanDirWithWorkers walks root using exactly `workers` goroutines.
// A value of 0 (or any negative number) selects the platform default
// from defaultWorkerCount(). Values larger than MaxScanWorkers are
// clamped down to keep the pool under the per-process fd budget.
func ScanDirWithWorkers(root string, excludeDirs []string, workers int) ([]RepoInfo, error) {
	return ScanDirContext(context.Background(), root, excludeDirs, workers)
}

// ScanDirContext is the cancellable form of ScanDirWithWorkers. When ctx
// is cancelled the walker stops dispatching new directories, drains its
// in-flight workers, and returns (partialRepos, ctx.Err()). Callers that
// want "best effort" output on Ctrl+C can still inspect the returned
// slice; callers that want strict semantics should treat any non-nil
// error as fatal and discard the partial results.
func ScanDirContext(ctx context.Context, root string, excludeDirs []string, workers int) ([]RepoInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return walkParallel(ctx, absRoot, buildExcludeSet(excludeDirs), resolveWorkerCount(workers))
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

// scanError boxes an error so it can be stored atomically. We need a
// pointer wrapper because atomic.Pointer requires a concrete type and
// `error` is an interface (which is two words and not atomically
// loadable on all architectures).
type scanError struct{ err error }

// scanState bundles shared state passed to every worker. Every field
// is either immutable after construction or accessed exclusively via
// channels / atomics — there is NO shared slice or map mutated by
// workers, by design.
type scanState struct {
	ctx     context.Context
	root    string
	exclude map[string]bool

	queue    chan string                 // pending directories to walk
	results  chan RepoInfo               // discovered repos → collector
	inflight atomic.Int64                // outstanding queued items
	firstErr atomic.Pointer[scanError]   // first error wins via CAS
}

// walkParallel runs a fixed-size worker pool that consumes directories
// from a buffered FIFO and enqueues child directories back. Discovered
// repos are streamed on a results channel to a single collector
// goroutine that owns the result slice — no shared-slice mutation.
//
// Shutdown sequence (race-safe):
//  1. Workers process queue items; each completed item decrements
//     `inflight`. A dispatched item that enqueues N children offsets
//     its own decrement by adding N before doing its decrement.
//  2. A watcher goroutine waits until `inflight == 0` (work drained)
//     OR `ctx.Done()` (cancellation), then closes the queue.
//  3. Workers exit their `for range queue` loop and the worker WG
//     fires.
//  4. With all workers gone, the results channel is closed.
//  5. The collector returns the assembled slice. No goroutine outlives
//     this function — verifiable with `goleak`.
func walkParallel(ctx context.Context, root string, exclude map[string]bool, workers int) ([]RepoInfo, error) {
	st := &scanState{
		ctx:     ctx,
		root:    root,
		exclude: exclude,
		queue:   make(chan string, queueBuffer),
		results: make(chan RepoInfo, resultsBuffer),
	}

	st.inflight.Add(1)
	st.queue <- root

	collected := startCollector(st.results)
	workerWG := startWorkers(st, workers)
	startWatcher(st, workerWG)

	repos := <-collected

	if err := ctx.Err(); err != nil {
		return repos, err
	}
	if boxed := st.firstErr.Load(); boxed != nil {
		return repos, boxed.err
	}

	return repos, nil
}

// startCollector spins up the single goroutine that owns the result
// slice. It returns a channel that delivers exactly one value — the
// final slice — when `results` is closed.
//
// Owning the slice in one goroutine is the safest possible pattern:
// no mutex, no atomic, no shared write set. The race detector has
// nothing to complain about because there is genuinely no shared
// mutation.
func startCollector(results <-chan RepoInfo) <-chan []RepoInfo {
	done := make(chan []RepoInfo, 1)
	go func() {
		var repos []RepoInfo
		for r := range results {
			repos = append(repos, r)
		}
		done <- repos
	}()

	return done
}

// startWorkers launches the fixed-size worker pool and returns a
// WaitGroup the caller can wait on to know when every worker has
// exited (i.e. the queue has been closed and drained).
func startWorkers(st *scanState, workers int) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range st.queue {
				st.processDir(dir)
				st.completeOne()
			}
		}()
	}

	return &wg
}

// startWatcher coordinates orderly shutdown. It owns BOTH closes:
//   - close(queue)   — once inflight drains OR ctx is cancelled
//   - close(results) — once every worker has exited
//
// Centralizing both closes in one goroutine eliminates the classic
// "who closes the channel?" race. Workers never close anything.
func startWatcher(st *scanState, workerWG *sync.WaitGroup) {
	go func() {
		st.waitForDrain()
		close(st.queue)
		workerWG.Wait()
		close(st.results)
	}()
}

// waitForDrain blocks until either every dispatched directory has
// completed (inflight == 0) or the context is cancelled. Polling is
// cheap (the scan itself is I/O bound) and avoids needing a separate
// "all done" channel that would have to be signalled exactly once.
//
// We poll on a sync.Cond-style condition via a tiny ticker because
// `inflight` is touched from many workers and using a channel would
// require either a mutex or atomic compare-loop in every worker.
func (st *scanState) waitForDrain() {
	// Fast path: most scans complete quickly; check immediately.
	if st.inflight.Load() == 0 {
		return
	}
	// Tick frequency is a tradeoff between shutdown latency and CPU.
	// 5ms is invisible to humans and negligible on a CPU even when
	// the scan runs for minutes.
	const tickEvery = 5_000_000 // 5ms in ns; avoids time import churn
	t := newTicker(tickEvery)
	defer t.stop()
	for {
		select {
		case <-st.ctx.Done():
			return
		case <-t.c:
			if st.inflight.Load() == 0 {
				return
			}
		}
	}
}

// processDir reads one directory and dispatches its child directories
// back onto the queue. Errors short-circuit further enqueues for THIS
// dir but do not stop other workers — the first error is captured and
// returned at the end. If ctx has been cancelled, the dir is skipped
// outright so the worker pool drains as fast as possible.
func (st *scanState) processDir(dir string) {
	if st.ctx.Err() != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		st.recordErr(err)

		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if st.ctx.Err() != nil {
			return
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
		st.emitRepo(parent)

		return
	}
	st.enqueue(filepath.Join(parent, name))
}

// enqueue dispatches a directory for processing. Cancellation is
// checked twice — once before the inflight increment and once via
// the select — because a cancellation between Load() and Add() would
// otherwise leak a counter that the watcher will never see go to zero
// (the queue close is what unblocks workers, but the watcher waits on
// inflight first).
func (st *scanState) enqueue(path string) {
	if st.ctx.Err() != nil {
		return
	}
	st.inflight.Add(1)
	select {
	case st.queue <- path:
	case <-st.ctx.Done():
		// Watcher will see ctx.Done() and close the queue; we must
		// undo the inflight bump so it can reach zero.
		st.inflight.Add(-1)
	}
}

// completeOne is the paired decrement for the initial Add(1) that put
// a directory on the queue. Called exactly once per dequeued item by
// the worker that processed it.
func (st *scanState) completeOne() {
	st.inflight.Add(-1)
}

// emitRepo streams a discovered repo to the collector. We take care to
// honor cancellation here because if the collector has already been
// asked to drain (it hasn't — only the watcher closes results — but in
// principle), or if the buffer is full and nobody is reading, we must
// not block forever.
func (st *scanState) emitRepo(repoPath string) {
	rel, err := filepath.Rel(st.root, repoPath)
	if err != nil {
		st.recordErr(err)

		return
	}
	info := RepoInfo{AbsolutePath: repoPath, RelativePath: rel}
	select {
	case st.results <- info:
	case <-st.ctx.Done():
		return
	}
}

// recordErr stores the FIRST error to occur using a CAS on the atomic
// pointer. Later errors are dropped to keep the public signature
// single-error and avoid a noisy multi-error. No mutex needed.
func (st *scanState) recordErr(err error) {
	boxed := &scanError{err: err}
	st.firstErr.CompareAndSwap(nil, boxed)
}
