package constants

// gitmap:cmd top-level
// `gitmap db-migrate` (alias `dbm`) — explicitly run database schema
// migrations. Safe to invoke at any time; designed to be run automatically
// after `gitmap update` and on first launch after a fresh install.
const (
	CmdDBMigrate      = "db-migrate"
	CmdDBMigrateAlias = "dbm"

	HelpDBMigrate = "  db-migrate (dbm)    Run pending database schema migrations (safe, idempotent)"

	FlagDBMigrateVerbose = "verbose"
	FlagDescDBMigrateV   = "Print every migration step (otherwise summary only)"

	// --force clears the persisted schema_version marker before Migrate()
	// runs, forcing the full v15 phase pipeline to re-execute even when the
	// fast-path would otherwise skip it. Useful when a previous run stamped
	// the marker but a downstream issue (corrupt seed, manual edit, partial
	// restore) means the schema actually needs re-walking — without paying
	// the full cost of `gitmap db-reset --confirm`.
	FlagDBMigrateForce = "force"
	FlagDescDBMigrateF = "Clear the schema_version marker first, forcing the full migration pipeline to re-run"

	MsgDBMigrateForceClear  = "  ▸ --force: cleared schema_version marker (full pipeline will re-run).\n"
	WarnDBMigrateForceClear = "  ⚠ --force: could not clear schema_version marker: %v\n"

	MsgDBMigrateRunning = "▸ Running gitmap database migrations...\n"
	MsgDBMigrateDoneFmt = "  ✓ Migrations complete (%d tables ensured, %d steps applied, %d warnings).\n"
	MsgDBMigrateStepFmt = "    • %s\n"
	MsgDBMigrateNoWork  = "  ✓ Database schema is already up to date — nothing to migrate.\n"
	ErrDBMigrateFailFmt = "Error: gitmap db-migrate failed: %v\n"

	// Auto-run hook for the post-update flow.
	MsgDBMigratePostUpdate = "▸ Running database migrations after update...\n"
	WarnDBMigratePostFail  = "  ⚠ Post-update migration failed: %v\n      Run `gitmap db-migrate --verbose` to retry.\n"
)
