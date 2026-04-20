package constants

import "testing"

// topLevelCmds enumerates every top-level Cmd* constant exposed to the CLI
// dispatcher. Entries marked with the `// gitmap:cmd skip` comment in
// constants_cli.go (subcommand verbs like "create" / "add" that are reused
// across subcommand groups) are intentionally omitted — duplicates of those
// values are expected and safe.
//
// When you add or remove a top-level Cmd* constant in constants_cli.go,
// update this slice. CI enforces parity via TestTopLevelCmd*.
func topLevelCmds() map[string]string {
	return map[string]string{
		"CmdScan":                  CmdScan,
		"CmdScanAlias":             CmdScanAlias,
		"CmdClone":                 CmdClone,
		"CmdCloneAlias":            CmdCloneAlias,
		"CmdUpdate":                CmdUpdate,
		"CmdInstalledDirAlias":     CmdInstalledDirAlias,
		"CmdVersion":               CmdVersion,
		"CmdVersionAlias":          CmdVersionAlias,
		"CmdHelp":                  CmdHelp,
		"CmdDesktopSync":           CmdDesktopSync,
		"CmdDesktopSyncAlias":      CmdDesktopSyncAlias,
		"CmdPull":                  CmdPull,
		"CmdPullAlias":             CmdPullAlias,
		"CmdRescan":                CmdRescan,
		"CmdRescanAlias":           CmdRescanAlias,
		"CmdSetup":                 CmdSetup,
		"CmdStatus":                CmdStatus,
		"CmdStatusAlias":           CmdStatusAlias,
		"CmdExec":                  CmdExec,
		"CmdExecAlias":             CmdExecAlias,
		"CmdRelease":               CmdRelease,
		"CmdReleaseShort":          CmdReleaseShort,
		"CmdReleaseBranch":         CmdReleaseBranch,
		"CmdReleaseBranchAlias":    CmdReleaseBranchAlias,
		"CmdReleasePending":        CmdReleasePending,
		"CmdReleasePendingAlias":   CmdReleasePendingAlias,
		"CmdChangelog":             CmdChangelog,
		"CmdChangelogAlias":        CmdChangelogAlias,
		"CmdDoctor":                CmdDoctor,
		"CmdLatestBranch":          CmdLatestBranch,
		"CmdLatestBranchAlias":     CmdLatestBranchAlias,
		"CmdList":                  CmdList,
		"CmdListAlias":             CmdListAlias,
		"CmdGroup":                 CmdGroup,
		"CmdGroupAlias":            CmdGroupAlias,
		"CmdDBReset":               CmdDBReset,
		"CmdReset":                 CmdReset,
		"CmdListVersions":          CmdListVersions,
		"CmdListVersionsAlias":     CmdListVersionsAlias,
		"CmdRevert":                CmdRevert,
		"CmdListReleases":          CmdListReleases,
		"CmdListReleasesAlias":     CmdListReleasesAlias,
		"CmdCompletion":            CmdCompletion,
		"CmdCompletionAlias":       CmdCompletionAlias,
		"CmdClearReleaseJSON":      CmdClearReleaseJSON,
		"CmdClearReleaseJSONAlias": CmdClearReleaseJSONAlias,
		"CmdDocs":                  CmdDocs,
		"CmdDocsAlias":             CmdDocsAlias,
		"CmdCloneNext":             CmdCloneNext,
		"CmdCloneNextAlias":        CmdCloneNextAlias,
		"CmdReleaseSelf":           CmdReleaseSelf,
		"CmdReleaseSelfAlias":      CmdReleaseSelfAlias,
		"CmdReleaseSelfAlias2":     CmdReleaseSelfAlias2,
		"CmdHelpDashboard":         CmdHelpDashboard,
		"CmdHelpDashboardAlias":    CmdHelpDashboardAlias,
		"CmdLLMDocs":               CmdLLMDocs,
		"CmdLLMDocsAlias":          CmdLLMDocsAlias,
		"CmdSelfInstall":           CmdSelfInstall,
		"CmdSelfUninstall":         CmdSelfUninstall,
		"CmdSf":                    CmdSf,
		"CmdProbe":                 CmdProbe,
	}
}

// TestTopLevelCmdConstantsAreUnique asserts that every top-level Cmd*
// constant has a distinct value, so CI rejects accidental redeclarations
// or value collisions (e.g. two constants both equal to "cd") before they
// reach the runtime dispatcher.
func TestTopLevelCmdConstantsAreUnique(t *testing.T) {
	seen := make(map[string]string, len(topLevelCmds()))
	for name, value := range topLevelCmds() {
		if prev, exists := seen[value]; exists {
			t.Errorf("duplicate top-level Cmd constant value %q: %s collides with %s", value, name, prev)
			continue
		}
		seen[value] = name
	}
}

// TestTopLevelCmdAliasesAreUnique asserts that every short alias (any
// top-level Cmd* value of length <= 2) is unique across the entire CLI
// surface. A future CmdFooAlias = "ls" would collide with CmdListAlias and
// be rejected here. Long-form command names are covered by the broader
// TestTopLevelCmdConstantsAreUnique check above; this test focuses
// specifically on the short-alias namespace where collisions are easiest
// to introduce by accident and hardest to spot in code review.
func TestTopLevelCmdAliasesAreUnique(t *testing.T) {
	const maxAliasLen = 2
	seen := make(map[string]string)
	for name, value := range topLevelCmds() {
		if len(value) == 0 || len(value) > maxAliasLen {
			continue
		}
		if prev, exists := seen[value]; exists {
			t.Errorf("duplicate short alias %q: %s collides with %s", value, name, prev)
			continue
		}
		seen[value] = name
	}
}
