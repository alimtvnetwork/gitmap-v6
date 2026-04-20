package constants

// Settings table (v15: Setting singular). The PK is Key (TEXT) so no
// {Table}Id rename is needed.
const TableSetting = "Setting"

// Legacy plural retained for migration detection.
const LegacyTableSettings = "Settings"

// SQL: create Setting table (v15).
const SQLCreateSetting = `CREATE TABLE IF NOT EXISTS Setting (
	Key   TEXT PRIMARY KEY,
	Value TEXT NOT NULL
)`

// SQL: setting operations (v15).
const (
	SQLUpsertSetting = `INSERT INTO Setting (Key, Value) VALUES (?, ?)
		ON CONFLICT(Key) DO UPDATE SET Value=excluded.Value`

	SQLSelectSetting = "SELECT Value FROM Setting WHERE Key = ?"

	SQLDeleteSetting = "DELETE FROM Setting WHERE Key = ?"
)

// SQL: reset (v15 + legacy).
const (
	SQLDropSetting  = "DROP TABLE IF EXISTS Setting"
	SQLDropSettings = "DROP TABLE IF EXISTS Settings" // legacy
)

// Settings keys.
const (
	SettingActiveGroup      = "active_group"
	SettingActiveMultiGroup = "active_multi_group"
	SettingSourceRepoPath   = "source_repo_path"
)

// Settings error messages.
const (
	ErrDBSettingUpsert = "failed to save setting: %v"
	ErrDBSettingQuery  = "failed to read setting: %v"
)
