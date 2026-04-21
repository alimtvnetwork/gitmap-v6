package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// newCancellableContext returns a context that is cancelled on the
// first SIGINT / SIGTERM. The returned cleanup function MUST be called
// (typically via defer) to stop the signal handler and release the
// goroutine — otherwise long-lived processes accumulate handlers.
//
// On the SECOND signal we exit hard with code 130 (128 + SIGINT). This
// matches conventional shell behavior and gives users an escape hatch
// when a graceful drain is taking too long.
func newCancellableContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "\n  ⚠ Interrupt received — stopping scan...")
			cancel()
		case <-ctx.Done():
			return
		}
		// Second signal: hard exit.
		select {
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "  ⚠ Second interrupt — exiting now.")
			os.Exit(130)
		case <-ctx.Done():
			return
		}
	}()

	cleanup := func() {
		signal.Stop(sigCh)
		cancel()
	}

	return ctx, cleanup
}
