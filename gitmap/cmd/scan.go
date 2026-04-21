package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/config"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/desktop"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/mapper"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/scanner"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/store"
)

// runScan handles the "scan" subcommand.
func runScan(args []string) {
	checkHelp("scan", args)
	f := ParseScanFlags(args)
	cfg, err := config.LoadFromFile(f.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrConfigLoad, f.ConfigPath, err)
		os.Exit(1)
	}
	cfg = config.MergeWithFlags(cfg, f.Mode, f.Output, f.OutputPath)
	cache := model.ScanCache{
		Dir: f.Dir, ConfigPath: f.ConfigPath, Mode: f.Mode, Output: f.Output,
		OutFile: f.OutFile, OutputPath: f.OutputPath,
		GithubDesktop: f.GHDesktop, OpenFolder: f.OpenFolder, Quiet: f.Quiet,
	}
	executeScan(f, cfg, cache)
}

// executeScan performs the directory scan and outputs results.
func executeScan(f ScanFlags, cfg model.Config, cache model.ScanCache) {
	absDir, err := filepath.Abs(f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrScanFailed, f.Dir, err)
		os.Exit(1)
	}

	// Enqueue scan as a pending task before execution.
	workDir, wdErr := os.Getwd()
	if wdErr != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not determine working directory: %v\n", wdErr)
	}
	cmdArgs := buildCommandArgs(append([]string{"scan"}, os.Args[2:]...))
	taskID, taskDB := createPendingTask(constants.TaskTypeScan, absDir, workDir, "scan", cmdArgs)
	if taskDB != nil {
		defer taskDB.Close()
	}

	// Install a Ctrl+C handler scoped to the directory walk. Once the
	// scan returns we tear it down so downstream steps (DB upsert,
	// project detection, etc.) keep using their normal exit semantics.
	ctx, cancel := newCancellableContext()
	repos, err := scanner.ScanDirContext(ctx, absDir, cfg.ExcludeDirs, workers)
	cancel()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			failPendingTask(taskDB, taskID, "scan cancelled by user")
			fmt.Fprintln(os.Stderr, "  ⚠ Scan cancelled — no artifacts written.")
			os.Exit(130)
		}
		failPendingTask(taskDB, taskID, fmt.Sprintf(constants.ErrScanFailed, absDir, err))
		fmt.Fprintf(os.Stderr, constants.ErrScanFailed, absDir, err)
		os.Exit(1)
	}
	records := mapper.BuildRecords(repos, cfg.DefaultMode, cfg.Notes)
	outputDir := resolveOutputDir(cfg.OutputDir, absDir)
	fmt.Printf(constants.MsgSectionArtifacts, outputDir)
	writeAllOutputs(records, outputDir, outFile, quiet)
	saveScanCache(outputDir, cache)
	fmt.Print(constants.MsgSectionDatabase)
	upsertToDB(records, outputDir)
	tagReposWithScanFolder(absDir, records, quiet)
	records = alignRecordsWithDB(records, outputDir)
	fmt.Print(constants.MsgSectionProjects)
	detected := detectAllProjects(records)
	writeProjectJSONFiles(detected, outputDir)
	upsertProjectsToDB(detected, records, outputDir)
	importReleases(absDir, outputDir)
	addToDesktop(records, ghDesktop)
	syncRecordsToVSCodePM(records, noVSCodeSync, noAutoTags)
	openOutputFolder(outputDir, openFolder)
	fmt.Print(constants.MsgSectionDone)

	// Mark scan task as completed after all steps succeed.
	completePendingTask(taskDB, taskID)
}

// tagReposWithScanFolder registers absDir as a ScanFolder and tags every
// just-scanned repo with the resulting ScanFolderId. Failures are reported
// to stderr but do NOT fail the scan — the underlying Repo rows still exist.
func tagReposWithScanFolder(absDir string, records []model.ScanRecord, quiet bool) {
	db, err := store.OpenDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrProbeOpenDB, err)
		return
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	folder, err := db.EnsureScanFolder(absDir, "", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	paths := make([]string, 0, len(records))
	for _, r := range records {
		paths = append(paths, r.AbsolutePath)
	}
	if err := db.TagReposByScanFolder(folder.ID, paths); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	if !quiet {
		fmt.Printf(constants.MsgScanFolderTagged, len(paths), folder.ID)
	}
}

// upsertToDB persists scan results into the SQLite database.
func upsertToDB(records []model.ScanRecord, outputDir string) {
	db, err := store.OpenDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgDBUpsertFailed, err)
		return
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgDBUpsertFailed, err)
		return
	}

	if err := db.UpsertRepos(records); err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgDBUpsertFailed, err)
		return
	}
	fmt.Printf(constants.MsgDBUpsertDone, len(records))
}

// alignRecordsWithDB rewrites record IDs to match persisted repo IDs by path.
func alignRecordsWithDB(records []model.ScanRecord, outputDir string) []model.ScanRecord {
	db, err := store.OpenDefault()
	if err != nil {
		return records
	}
	defer db.Close()

	repos, err := db.ListRepos()
	if err != nil {
		return records
	}

	idsByPath := make(map[string]int64, len(repos))
	for _, repo := range repos {
		idsByPath[repo.AbsolutePath] = repo.ID
	}

	aligned := make([]model.ScanRecord, 0, len(records))
	for _, rec := range records {
		if id, ok := idsByPath[rec.AbsolutePath]; ok {
			rec.ID = id
		}
		aligned = append(aligned, rec)
	}

	return aligned
}

// addToDesktop registers repos with GitHub Desktop if requested.
func addToDesktop(records []model.ScanRecord, enabled bool) {
	if enabled {
		summary := desktop.AddRepos(records)
		fmt.Printf(constants.MsgDesktopSummary, summary.Added, summary.Failed)
	}
}

// openOutputFolder opens the output directory in the OS file explorer.
func openOutputFolder(outputDir string, enabled bool) {
	if enabled {
		cmd := resolveOpenCommand(outputDir)
		_ = cmd.Start()
		fmt.Printf(constants.MsgOpenedFolder, outputDir)
	}
}

// resolveOpenCommand returns the OS-specific command to open a folder.
func resolveOpenCommand(dir string) *exec.Cmd {
	if runtime.GOOS == constants.OSWindows {
		return exec.Command(constants.CmdExplorer, dir)
	}
	if runtime.GOOS == constants.OSDarwin {
		return exec.Command(constants.CmdOpen, dir)
	}

	return exec.Command(constants.CmdXdgOpen, dir)
}

// resolveOutputDir determines the output directory relative to scan root.
func resolveOutputDir(cfgDir, scanDir string) string {
	if filepath.IsAbs(cfgDir) {
		return cfgDir
	}

	return filepath.Join(scanDir, constants.GitMapDir, constants.OutputDirName)
}
