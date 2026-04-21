package cmd

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/lockfile"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/scripts"
)

// selfInstallOpts holds parsed flags for self-install.
type selfInstallOpts struct {
	Dir       string
	Yes       bool
	Version   string
	Profile   string // --profile: auto|both|zsh|bash|pwsh|fish (default auto)
	ShowPath  bool   // --show-path: expand install summary with PATH audit trail
	ForceLock bool   // --force-lock: bypass the duplicate-install guard
}

// runSelfInstall is the entry point for `gitmap self-install`. It picks
// the install directory (prompting with a default), then runs the
// embedded install script, falling back to GitHub if missing.
//
// A process-wide lock (gitmap-selfinstall.lock in os.TempDir) prevents
// two concurrent invocations from racing — otherwise the user sees the
// install prompt twice and PATH/binary writes overlap. See
// gitmap/lockfile and the discovery-delegation flow in scripts/install.sh.
func runSelfInstall(args []string) {
	checkHelp(constants.CmdSelfInstall, args)
	opts := parseSelfInstallFlags(args)
	release := acquireSelfInstallLock(opts)
	defer release()

	dir := resolveSelfInstallDir(opts)
	fmt.Print(constants.MsgSelfInstallHeader)
	fmt.Printf(constants.MsgSelfInstallUsing, dir)
	scriptName, scriptBody := loadInstallScript()
	tmpPath := writeInstallScriptTemp(scriptName, scriptBody)
	defer os.Remove(tmpPath)
	executeInstallScript(scriptName, tmpPath, dir, opts)
	fmt.Print(constants.MsgSelfInstallDone)
	fmt.Print(constants.MsgSelfInstallReminder)
}

// acquireSelfInstallLock takes the duplicate-install guard. On conflict
// the process exits 1 with a clear pointer to the holder's PID so the
// user knows which terminal/script is already installing. --force-lock
// skips the guard for stale-lock recovery.
func acquireSelfInstallLock(opts selfInstallOpts) lockfile.Releaser {
	if opts.ForceLock {
		release, err := lockfile.ForceAcquire(constants.SelfInstallLockName)
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrSelfInstallLock, err)
			os.Exit(1)
		}

		return release
	}
	release, err := lockfile.Acquire(constants.SelfInstallLockName)
	if err == nil {
		return release
	}
	if errors.Is(err, lockfile.ErrAlreadyHeld) {
		holder := lockfile.HolderPID(constants.SelfInstallLockName)
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallLockHeld, holder)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, constants.ErrSelfInstallLock, err)
	os.Exit(1)

	return func() {} // unreachable; satisfies the compiler
}

// parseSelfInstallFlags reads --dir / --yes / --version / --profile / --dual-shell / --show-path / --force-lock.
//
// --dual-shell is kept as a hidden alias for --profile both: when both
// flags are passed, --profile wins (explicit beats deprecated). This
// preserves backward compat for existing scripts/CI without forcing a
// breaking change.
func parseSelfInstallFlags(args []string) selfInstallOpts {
	fs := flag.NewFlagSet(constants.CmdSelfInstall, flag.ExitOnError)
	opts := selfInstallOpts{}
	var dualShell bool
	fs.StringVar(&opts.Dir, "dir", "", constants.FlagDescSelfDir)
	fs.BoolVar(&opts.Yes, "yes", false, constants.FlagDescSelfYes)
	fs.BoolVar(&opts.Yes, "y", false, constants.FlagDescSelfYes)
	fs.StringVar(&opts.Version, "version", "", constants.FlagDescSelfFromVersion)
	fs.StringVar(&opts.Profile, "profile", constants.ProfileModeAuto, constants.FlagDescSelfProfile)
	fs.BoolVar(&dualShell, "dual-shell", false, constants.FlagDescSelfDualShell)
	fs.BoolVar(&opts.ShowPath, "show-path", false, constants.FlagDescSelfShowPath)
	fs.BoolVar(&opts.ForceLock, "force-lock", false, constants.FlagDescSelfForceLock)
	fs.Parse(reorderFlagsBeforeArgs(args))
	opts.Profile = resolveProfileMode(opts.Profile, dualShell)
	validateProfileMode(opts.Profile)

	return opts
}

// resolveProfileMode reconciles --profile and the deprecated --dual-shell
// alias. Explicit --profile always wins; --dual-shell only upgrades the
// default `auto` to `both` so it never silently overrides a user choice.
func resolveProfileMode(profile string, dualShell bool) string {
	if dualShell && profile == constants.ProfileModeAuto {
		return constants.ProfileModeBoth
	}

	return profile
}

// validateProfileMode rejects unknown --profile values with a clear list
// of accepted ones. Exits 1 — bad CLI input is unrecoverable.
func validateProfileMode(mode string) {
	for _, valid := range constants.SelfInstallProfileModes {
		if mode == valid {
			return
		}
	}
	fmt.Fprintf(os.Stderr, constants.ErrSelfInstallProfileInvalid,
		mode, strings.Join(constants.SelfInstallProfileModes, "|"))
	os.Exit(1)
}

// resolveSelfInstallDir returns the install directory, prompting the
// user with a default if neither --dir nor --yes was supplied.
func resolveSelfInstallDir(opts selfInstallOpts) string {
	if len(opts.Dir) > 0 {
		return opts.Dir
	}
	def := defaultSelfInstallDir()
	if opts.Yes {
		return def
	}

	return promptInstallDir(def)
}

// defaultSelfInstallDir returns the platform-default install directory.
func defaultSelfInstallDir() string {
	if runtime.GOOS == "windows" {
		return constants.SelfInstallDefaultWindows
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin/gitmap"
	}

	return filepath.Join(home, constants.SelfInstallDefaultUnix)
}

// promptInstallDir asks the user for a path, accepting the default on
// empty input.
func promptInstallDir(def string) string {
	fmt.Printf(constants.MsgSelfInstallPrompt, def)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallReadStdin, err)
		os.Exit(1)
	}
	answer := strings.TrimSpace(line)
	if len(answer) == 0 {
		return def
	}

	return answer
}

// loadInstallScript returns the script name + bytes for the platform,
// preferring embedded copies and falling back to remote download.
func loadInstallScript() (string, []byte) {
	name := pickInstallScriptName()
	body, err := scripts.ReadFile(name)
	if err == nil && len(body) > 0 {
		fmt.Printf(constants.MsgSelfInstallEmbedded, name)

		return name, body
	}
	remote := pickInstallScriptURL()
	fmt.Printf(constants.MsgSelfInstallRemote, remote)
	body, dlErr := downloadInstallScript(remote)
	if dlErr != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallDownload, remote, dlErr)
		os.Exit(1)
	}

	return name, body
}

// pickInstallScriptName returns install.ps1 on Windows, install.sh elsewhere.
func pickInstallScriptName() string {
	if runtime.GOOS == "windows" {
		return constants.SelfInstallScriptPwsh
	}

	return constants.SelfInstallScriptBash
}

// pickInstallScriptURL is the GitHub fallback for the platform script.
func pickInstallScriptURL() string {
	if runtime.GOOS == "windows" {
		return constants.SelfInstallRemotePwsh
	}

	return constants.SelfInstallRemoteBash
}

// downloadInstallScript fetches an installer over HTTPS.
// URL is sourced from compile-time constants (SelfInstallRemotePwsh /
// SelfInstallRemoteBash); not user-controlled.
func downloadInstallScript(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // G107: URL is a compile-time constant.
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// writeInstallScriptTemp persists the install script to a temp file
// (with a UTF-8 BOM on PowerShell) so it can be invoked.
func writeInstallScriptTemp(name string, body []byte) string {
	pattern := "gitmap-self-install-*"
	if strings.HasSuffix(name, ".ps1") {
		pattern += ".ps1"
	} else {
		pattern += ".sh"
	}
	f, err := os.CreateTemp(os.TempDir(), pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallScriptWrite, err)
		os.Exit(1)
	}
	defer f.Close()
	if strings.HasSuffix(name, ".ps1") {
		if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrSelfInstallScriptWrite, err)
			os.Exit(1)
		}
	}
	if _, err := f.Write(body); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallScriptWrite, err)
		os.Exit(1)
	}
	if !strings.HasSuffix(name, ".ps1") {
		_ = os.Chmod(f.Name(), 0o755)
	}

	return f.Name()
}

// executeInstallScript invokes PowerShell or bash on the script with the
// resolved install directory and optional version / dual-shell mode.
func executeInstallScript(name, path, dir string, opts selfInstallOpts) {
	cmd := buildSelfInstallCmd(name, path, dir, opts)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSelfInstallScriptRun, err)
		os.Exit(1)
	}
}

// buildSelfInstallCmd assembles the platform-specific exec.Cmd. On Unix,
// when --dual-shell is set, GITMAP_DUAL_SHELL=1 is exported into the
// child's env so detect_active_pwsh fires even without other signals.
func buildSelfInstallCmd(name, path, dir string, opts selfInstallOpts) *exec.Cmd {
	if strings.HasSuffix(name, ".ps1") {
		return buildSelfInstallPwshCmd(path, dir, opts)
	}

	return buildSelfInstallBashCmd(path, dir, opts)
}

// buildSelfInstallPwshCmd builds the Windows / pwsh invocation.
// --dual-shell is currently a no-op on Windows (single-shell platform);
// kept in the arg list for forward compatibility with future PSCore-on-
// Linux / pwsh-side dual-write logic.
func buildSelfInstallPwshCmd(path, dir string, opts selfInstallOpts) *exec.Cmd {
	args := []string{"-ExecutionPolicy", "Bypass", "-NoProfile",
		"-NoLogo", "-File", path, "-InstallDir", dir}
	if len(opts.Version) > 0 {
		args = append(args, "-Version", opts.Version)
	}

	return exec.Command("pwsh", args...)
}

// buildSelfInstallBashCmd builds the Unix invocation and propagates
// --profile + --show-path through to install.sh. When the resolved
// profile mode is `both`, GITMAP_DUAL_SHELL=1 is also exported as a
// belt-and-suspenders signal so detect_active_pwsh inside install.sh
// fires from either the env var OR the explicit flag.
func buildSelfInstallBashCmd(path, dir string, opts selfInstallOpts) *exec.Cmd {
	args := []string{path, "--dir", dir}
	if len(opts.Version) > 0 {
		args = append(args, "--version", opts.Version)
	}
	args = append(args, constants.FlagSelfProfile, opts.Profile)
	if opts.ShowPath {
		args = append(args, constants.FlagSelfShowPath)
	}
	cmd := exec.Command("bash", args...)
	if opts.Profile == constants.ProfileModeBoth {
		cmd.Env = append(os.Environ(), "GITMAP_DUAL_SHELL=1")
	}

	return cmd
}
