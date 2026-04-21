package scanner

import "time"

// ticker is a tiny shim over time.Ticker so the scanner core file
// doesn't need to import "time" just for one usage site. Keeping the
// import surface small makes the concurrency code easier to audit.
type ticker struct {
	t *time.Ticker
	c <-chan time.Time
}

// newTicker creates a ticker that fires every `everyNanos` nanoseconds.
func newTicker(everyNanos int64) *ticker {
	t := time.NewTicker(time.Duration(everyNanos))

	return &ticker{t: t, c: t.C}
}

// stop halts the underlying ticker; safe to call multiple times.
func (k *ticker) stop() {
	k.t.Stop()
}
