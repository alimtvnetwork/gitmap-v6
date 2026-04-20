// Package cmd — scanprojects.go handles project detection during scan.
package cmd

import (
	"fmt"
	"os"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/detector"
	"github.com/user/gitmap/model"
	"github.com/user/gitmap/store"
)

// detectAllProjects runs project detection across all scanned repos.
func detectAllProjects(records []model.ScanRecord) []detector.DetectionResult {
	var all []detector.DetectionResult
	repoCount := 0
	for _, rec := range records {
		results := detector.DetectProjects(rec.AbsolutePath, rec.ID, rec.RepoName)
		if len(results) > 0 {
			repoCount++
			all = append(all, results...)
		}
	}
	fmt.Printf(constants.MsgProjectDetectDone, len(all), repoCount)

	return all
}

// upsertProjectsToDB persists detected projects and metadata to SQLite.
func upsertProjectsToDB(results []detector.DetectionResult, records []model.ScanRecord, outputDir string) {
	if len(results) == 0 {
		return
	}
	db, err := store.OpenDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrProjectUpsert, err)

		return
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrProjectUpsert, err)

		return
	}
	upsertProjectRecords(db, results, records)
}

// upsertProjectRecords inserts projects, metadata, and cleans stale records.
func upsertProjectRecords(db *store.DB, results []detector.DetectionResult, records []model.ScanRecord) {
	count := 0
	repoIDs := collectRepoIDs(results)
	for i := range results {
		r := &results[i]
		err := db.UpsertDetectedProject(r.Project)
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrProjectUpsert, err)

			continue
		}
		if err := resolveDetectedProjectID(db, r); err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrProjectUpsert, err)

			continue
		}
		count++
		upsertProjectMetadata(db, *r)
	}
	cleanStaleProjects(db, repoIDs, results)
	fmt.Printf(constants.MsgProjectUpsertDone, count)
}

// resolveDetectedProjectID syncs the project ID with the persisted DB row.
func resolveDetectedProjectID(db *store.DB, r *detector.DetectionResult) error {
	id, err := db.SelectDetectedProjectID(
		r.Project.RepoID,
		r.Project.ProjectTypeID,
		r.Project.RelativePath,
	)
	if err != nil {
		return err
	}
	r.Project.ID = id

	return nil
}

// upsertProjectMetadata persists Go or C# metadata for a detection result.
func upsertProjectMetadata(db *store.DB, r detector.DetectionResult) {
	if r.GoMeta != nil {
		upsertGoProjectMeta(db, r)
	}
	if r.Csharp != nil {
		upsertCsharpProjectMeta(db, r)
	}
}
