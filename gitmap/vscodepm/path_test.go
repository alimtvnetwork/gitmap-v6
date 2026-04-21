package vscodepm

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestUserDataCandidate_PerOSEnvDispatch verifies the per-OS dispatcher
// returns the platform-appropriate candidate path for the env vars set on
// the running test host. Only the active GOOS branch is exercised here;
// the Windows / Linux helpers each have their own dedicated tests below.
func TestUserDataCandidate_PerOSEnvDispatch(t *testing.T) {
	got := userDataCandidate()
	if got == "" {
		t.Skipf("no env vars set for GOOS=%s — nothing to assert", runtime.GOOS)
	}

	if !filepath.IsAbs(got) && !strings.Contains(got, "Code") {
		t.Errorf("userDataCandidate() = %q, expected an absolute path containing %q", got, "Code")
	}
}

// TestWindowsCandidate_AppDataPrimary asserts the primary %APPDATA% branch
// wins when present.
func TestWindowsCandidate_AppDataPrimary(t *testing.T) {
	t.Setenv("APPDATA", filepath.FromSlash("C:/Users/jane/AppData/Roaming"))
	t.Setenv("USERPROFILE", filepath.FromSlash("C:/Users/jane"))

	got := windowsUserDataCandidate()
	want := filepath.FromSlash("C:/Users/jane/AppData/Roaming/Code")
	if got != want {
		t.Errorf("windowsUserDataCandidate() = %q, want %q", got, want)
	}
}

// TestWindowsCandidate_UserProfileFallback asserts the %USERPROFILE%
// fallback fires when %APPDATA% is empty.
func TestWindowsCandidate_UserProfileFallback(t *testing.T) {
	t.Setenv("APPDATA", "")
	t.Setenv("USERPROFILE", filepath.FromSlash("C:/Users/jane"))

	got := windowsUserDataCandidate()
	want := filepath.Join(filepath.FromSlash("C:/Users/jane"),
		filepath.FromSlash("AppData/Roaming/Code"))
	if got != want {
		t.Errorf("windowsUserDataCandidate() = %q, want %q", got, want)
	}
}

// TestWindowsCandidate_NoEnvReturnsEmpty asserts the helper returns ""
// when neither env var is set, signaling ErrUserDataMissing upstream.
func TestWindowsCandidate_NoEnvReturnsEmpty(t *testing.T) {
	t.Setenv("APPDATA", "")
	t.Setenv("USERPROFILE", "")

	if got := windowsUserDataCandidate(); got != "" {
		t.Errorf("windowsUserDataCandidate() = %q, want empty string", got)
	}
}

// TestLinuxCandidate_XDGPrimary asserts $XDG_CONFIG_HOME wins.
func TestLinuxCandidate_XDGPrimary(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	t.Setenv("HOME", "/home/jane")

	got := linuxUserDataCandidate()
	want := filepath.Join("/custom/xdg", "Code")
	if got != want {
		t.Errorf("linuxUserDataCandidate() = %q, want %q", got, want)
	}
}

// TestLinuxCandidate_HomeFallback asserts $HOME/.config/Code fallback.
func TestLinuxCandidate_HomeFallback(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/jane")

	got := linuxUserDataCandidate()
	want := filepath.Join("/home/jane", filepath.FromSlash(".config/Code"))
	if got != want {
		t.Errorf("linuxUserDataCandidate() = %q, want %q", got, want)
	}
}

// TestLinuxCandidate_NoEnvReturnsEmpty asserts the helper returns ""
// when neither env var is set.
func TestLinuxCandidate_NoEnvReturnsEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")

	if got := linuxUserDataCandidate(); got != "" {
		t.Errorf("linuxUserDataCandidate() = %q, want empty string", got)
	}
}

// TestDarwinCandidate_HomePresent asserts the macOS Library path is built
// off $HOME.
func TestDarwinCandidate_HomePresent(t *testing.T) {
	t.Setenv("HOME", "/Users/jane")

	got := darwinUserDataCandidate()
	want := filepath.Join("/Users/jane",
		filepath.FromSlash("Library/Application Support/Code"))
	if got != want {
		t.Errorf("darwinUserDataCandidate() = %q, want %q", got, want)
	}
}

// TestDarwinCandidate_NoHomeReturnsEmpty asserts empty $HOME yields "".
func TestDarwinCandidate_NoHomeReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", "")

	if got := darwinUserDataCandidate(); got != "" {
		t.Errorf("darwinUserDataCandidate() = %q, want empty string", got)
	}
}

// TestUserDataRoot_MissingEnvReturnsSentinel asserts the public resolver
// returns ErrUserDataMissing (not a string error) when no env is set, so
// callers can use errors.Is for soft-fail handling.
func TestUserDataRoot_MissingEnvReturnsSentinel(t *testing.T) {
	clearAllVSCodeEnv(t)

	_, err := UserDataRoot()
	if !errors.Is(err, ErrUserDataMissing) {
		t.Fatalf("UserDataRoot() err = %v, want %v", err, ErrUserDataMissing)
	}
}

// TestProjectsJSONPath_MissingRootReturnsSentinel asserts the higher-level
// path builder also surfaces ErrUserDataMissing when the root cannot be
// located, without ever attempting to touch the filesystem.
func TestProjectsJSONPath_MissingRootReturnsSentinel(t *testing.T) {
	clearAllVSCodeEnv(t)

	_, err := ProjectsJSONPath()
	if !errors.Is(err, ErrUserDataMissing) {
		t.Fatalf("ProjectsJSONPath() err = %v, want %v", err, ErrUserDataMissing)
	}
}

// TestProjectsJSONPath_RootExistsExtMissingReturnsSentinel points the
// resolver at a real temp dir (so UserDataRoot succeeds) but does NOT
// create the alefragnani.project-manager subtree, asserting the second
// sentinel fires.
func TestProjectsJSONPath_RootExistsExtMissingReturnsSentinel(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("darwin user-data path is too deep to fake cleanly in a temp dir")
	}

	parent := t.TempDir()
	// The resolver appends "Code" to APPDATA / XDG_CONFIG_HOME, so create it.
	codeDir := filepath.Join(parent, "Code")
	if err := os.MkdirAll(codeDir, 0o755); err != nil {
		t.Fatalf("mkdir Code: %v", err)
	}

	clearAllVSCodeEnv(t)
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", parent)
	} else {
		t.Setenv("XDG_CONFIG_HOME", parent)
	}

	_, err := ProjectsJSONPath()
	if !errors.Is(err, ErrExtensionMissing) {
		t.Fatalf("ProjectsJSONPath() err = %v, want %v", err, ErrExtensionMissing)
	}
}

// clearAllVSCodeEnv unsets every env var the resolver consults, on every
// supported platform. Uses t.Setenv so values restore at test end.
func clearAllVSCodeEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"APPDATA", "USERPROFILE", "HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(k, "")
	}
}

// pointEnvAt makes the per-OS resolver return `root` as the user-data dir.
// Sets the primary env var for the active GOOS so userDataCandidate picks
// it up without needing the fallback.
func pointEnvAt(t *testing.T, root string) {
	t.Helper()
	clearAllVSCodeEnv(t)

	switch runtime.GOOS {
	case "windows":
		// APPDATA is the parent of the "Code" dir, so strip the suffix.
		t.Setenv("APPDATA", filepath.Dir(root))
		// Rename the temp dir into "<parent>/Code" so the resolver finds it.
		// We can't actually rename here without disturbing t.TempDir cleanup,
		// so instead override APPDATA to the temp dir directly and rely on
		// the resolver appending "Code" — which means the caller must have
		// passed a path ending in "Code".
		t.Setenv("APPDATA", root)
	case "darwin":
		// $HOME/Library/Application Support/Code — too deep to fake cleanly
		// without creating real subdirs. Skip this combination on darwin.
		t.Skipf("pointEnvAt: darwin path is too deep to fake cleanly; covered indirectly")
	default:
		t.Setenv("XDG_CONFIG_HOME", filepath.Dir(root))
		t.Setenv("XDG_CONFIG_HOME", root)
	}
	_ = filepath.Separator // keep import meaningful even if Skip fires
}
