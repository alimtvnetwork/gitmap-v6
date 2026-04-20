package store

import (
	"fmt"
	"strings"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/model"
)

// UpsertRelease inserts or updates a release record in the database.
// v15: persists IsDraft / IsPreRelease columns.
func (db *DB) UpsertRelease(r model.ReleaseRecord) error {
	isDraft := boolToInt(r.IsDraft)
	isPreRelease := boolToInt(r.IsPreRelease)
	isLatest := boolToInt(r.IsLatest)

	if r.IsLatest {
		if err := db.clearLatest(); err != nil {
			return err
		}
	}

	_, err := db.conn.Exec(constants.SQLUpsertRelease,
		r.Version, r.Tag, r.Branch, r.SourceBranch,
		r.CommitSha, r.Changelog, r.Notes, isDraft, isPreRelease, isLatest, r.Source, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf(constants.ErrDBReleaseUpsert, err)
	}

	return nil
}

// ListReleases returns all releases ordered by creation date descending.
func (db *DB) ListReleases() ([]model.ReleaseRecord, error) {
	rows, err := db.conn.Query(constants.SQLSelectAllReleases)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrDBReleaseQuery, err)
	}
	defer rows.Close()

	return scanReleaseRows(rows)
}

// FindReleaseByTag returns a release matching the given tag.
func (db *DB) FindReleaseByTag(tag string) (model.ReleaseRecord, error) {
	row := db.conn.QueryRow(constants.SQLSelectReleaseByTag, tag)

	return scanOneRelease(row)
}

// clearLatest resets the IsLatest flag on all existing releases.
func (db *DB) clearLatest() error {
	_, err := db.conn.Exec(constants.SQLClearLatestRelease)
	if err != nil {
		return fmt.Errorf(constants.ErrDBReleaseUpsert, err)
	}

	return nil
}

// scanReleaseRows reads ReleaseRecord values from query result rows.
func scanReleaseRows(rows interface {
	Next() bool
	Scan(dest ...any) error
}) ([]model.ReleaseRecord, error) {
	var results []model.ReleaseRecord

	for rows.Next() {
		r, err := scanOneReleaseRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

// scanOneReleaseRow reads a single ReleaseRecord from a row scanner.
func scanOneReleaseRow(row interface{ Scan(dest ...any) error }) (model.ReleaseRecord, error) {
	var r model.ReleaseRecord
	var isDraft, isPreRelease, isLatest int

	err := row.Scan(&r.ID, &r.Version, &r.Tag, &r.Branch, &r.SourceBranch,
		&r.CommitSha, &r.Changelog, &r.Notes, &isDraft, &isPreRelease, &isLatest, &r.Source, &r.CreatedAt)
	if err != nil {
		return model.ReleaseRecord{}, err
	}

	r.IsDraft = isDraft == 1
	r.IsPreRelease = isPreRelease == 1
	r.IsLatest = isLatest == 1

	return r, nil
}

// scanOneRelease reads a single ReleaseRecord from a QueryRow result.
func scanOneRelease(row interface{ Scan(dest ...any) error }) (model.ReleaseRecord, error) {
	return scanOneReleaseRow(row)
}

// JoinChangelog joins changelog notes into a newline-separated string.
func JoinChangelog(notes []string) string {
	if len(notes) == 0 {
		return ""
	}

	return strings.Join(notes, "\n")
}

// boolToInt converts a bool to SQLite integer (0/1).
func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
