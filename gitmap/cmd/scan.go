package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/user/gitmap/config"
	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/desktop"

	"github.com/user/gitmap/mapper"
	"github.com/user/gitmap/model"
	"github.com/user/gitmap/scanner"
	"github.com/user/gitmap/store"
)

// runScan handles the "scan" subcommand.
func runScan(args []string) {
	checkHelp("scan", args)
	dir, cfgPath, mode, output, outFile, outputPath, ghDesktop, openFolder, quiet := parseScanFlags(args)
	cfg, err := config.LoadFromFile(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrConfigLoad, cfgPath, err)
		os.Exit(1)
	}
	cfg = config.MergeWithFlags(cfg, mode, output, outputPath)
	cache := model.ScanCache{
		Dir: dir, ConfigPath: cfgPath, Mode: mode, Output: output,
		OutFile: outFile, OutputPath: outputPath,
		GithubDesktop: ghDesktop, OpenFolder: openFolder, Quiet: quiet,
	}
	executeScan(dir, cfg, outFile, ghDesktop, openFolder, quiet, cache)
}

// executeScan performs the directory scan and outputs results.
func executeScan(dir string, cfg model.Config, outFile string, ghDesktop, openFolder, quiet bool, cache model.ScanCache) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrScanFailed, dir, err)
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

	repos, err := scanner.ScanDir(absDir, cfg.ExcludeDirs)
	if err != nil {
		failPendingTask(taskDB, taskID, fmt.Sprintf(constants.ErrScanFailed, absDir, err))
		fmt.Fprintf(os.Stderr, constants.ErrScanFailed, absDir, err)
		os.Exit(1)
	}
	records := mapper.BuildRecords(repos, cfg.DefaultMode, cfg.Notes)
	outputDir := resolveOutputDir(cfg.OutputDir, absDir)
	writeAllOutputs(records, outputDir, outFile, quiet)
	saveScanCache(outputDir, cache)
	upsertToDB(records, outputDir)
	tagReposWithScanFolder(absDir, records, quiet)
	records = alignRecordsWithDB(records, outputDir)
	detected := detectAllProjects(records)
	writeProjectJSONFiles(detected, outputDir)
	upsertProjectsToDB(detected, records, outputDir)
	importReleases(absDir, outputDir)
	addToDesktop(records, ghDesktop)
	openOutputFolder(outputDir, openFolder)

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
		fmt.Printf("✓ Tagged %d repo(s) with scan folder #%d\n", len(paths), folder.ID)
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
