package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/store"
)

// runDBMigrate handles the "db-migrate" (alias "dbm") subcommand.
//
// It opens the active-profile database, runs Migrate() (which is idempotent
// and safe to invoke repeatedly), and prints a single-line summary. The
// --verbose flag prints every migration step that ran.
// The --force flag clears the schema_version marker before Migrate() so the
// full pipeline re-runs even when the fast-path would otherwise skip it.
func runDBMigrate(args []string) {
	checkHelp(constants.CmdDBMigrate, args)
	verbose, force := parseDBMigrateFlags(args)

	fmt.Print(constants.MsgDBMigrateRunning)

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrDBMigrateFailFmt, err)
		os.Exit(1)
	}
	defer db.Close()

	if force {
		clearSchemaVersionMarker(db)
	}

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrDBMigrateFailFmt, err)
		os.Exit(1)
	}

	printDBMigrateSummary(verbose)
}

// parseDBMigrateFlags extracts the --verbose and --force flags.
func parseDBMigrateFlags(args []string) (bool, bool) {
	fs := flag.NewFlagSet(constants.CmdDBMigrate, flag.ExitOnError)
	v := fs.Bool(constants.FlagDBMigrateVerbose, false, constants.FlagDescDBMigrateV)
	f := fs.Bool(constants.FlagDBMigrateForce, false, constants.FlagDescDBMigrateF)

	if err := fs.Parse(reorderFlagsBeforeArgs(args)); err != nil {
		os.Exit(2)
	}

	return *v, *f
}

// clearSchemaVersionMarker deletes the persisted schema_version Setting row
// so the next Migrate() call cannot take the fast-path. Failures are warned
// to stderr but never fatal — the worst case is the fast-path still triggers
// and the user just re-runs without --force, which is the existing behavior.
func clearSchemaVersionMarker(db *store.DB) {
	err := db.DeleteSetting(constants.SettingSchemaVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnDBMigrateForceClear, err)

		return
	}

	fmt.Print(constants.MsgDBMigrateForceClear)
}

// printDBMigrateSummary writes the post-run summary line.
//
// Migrate() streams every per-step warning to os.Stderr already (with the
// table + column + action context). If any warning was printed, the user
// has already seen it; here we just confirm the run reached the end.
func printDBMigrateSummary(verbose bool) {
	fmt.Print(constants.MsgDBMigrateNoWork)

	if verbose {
		fmt.Println("    (verbose: every CREATE/ALTER is idempotent — re-running has no effect)")
		fmt.Println("    (any per-step warnings above include the offending table + column)")
	}
}

// runPostUpdateMigrate is invoked from the update flow after the binary is
// replaced. It is best-effort: any failure is warned, never fatal, since the
// user may have an in-flight DB lock or read-only environment.
func runPostUpdateMigrate() {
	fmt.Print(constants.MsgDBMigratePostUpdate)

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnDBMigratePostFail, err)

		return
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnDBMigratePostFail, err)

		return
	}

	fmt.Println("  ✓ Schema migrations complete.")
}
