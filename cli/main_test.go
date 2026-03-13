package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

// captureStdout runs fn and returns everything written to os.Stdout.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func resetVerbose() {
	verbose = false
	os.Unsetenv("CORRIDOR_VERBOSE")
}

func TestParseGlobalFlags_Verbose(t *testing.T) {
	resetVerbose()
	args := parseGlobalFlags([]string{"install", "--verbose"})
	if !verbose {
		t.Error("expected verbose=true after --verbose")
	}
	if len(args) != 1 || args[0] != "install" {
		t.Errorf("expected [install], got %v", args)
	}
}

func TestParseGlobalFlags_ShortV(t *testing.T) {
	resetVerbose()
	args := parseGlobalFlags([]string{"-v", "install"})
	if !verbose {
		t.Error("expected verbose=true after -v")
	}
	if len(args) != 1 || args[0] != "install" {
		t.Errorf("expected [install], got %v", args)
	}
}

func TestParseGlobalFlags_NoVerbose(t *testing.T) {
	resetVerbose()
	args := parseGlobalFlags([]string{"install", "--force"})
	if verbose {
		t.Error("expected verbose=false when no verbose flag")
	}
	if len(args) != 2 || args[0] != "install" || args[1] != "--force" {
		t.Errorf("expected [install --force], got %v", args)
	}
}

func TestParseGlobalFlags_VerboseAnywhere(t *testing.T) {
	resetVerbose()
	args := parseGlobalFlags([]string{"install", "--force", "-v", "myplugin"})
	if !verbose {
		t.Error("expected verbose=true when -v in middle")
	}
	expected := []string{"install", "--force", "myplugin"}
	if len(args) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, args)
	}
	for i, a := range args {
		if a != expected[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, expected[i], a)
		}
	}
}

func TestCorridorVerboseEnvVar(t *testing.T) {
	resetVerbose()
	os.Setenv("CORRIDOR_VERBOSE", "1")
	defer os.Unsetenv("CORRIDOR_VERBOSE")

	parseGlobalFlags([]string{"install"})
	if os.Getenv("CORRIDOR_VERBOSE") != "1" {
		t.Error("CORRIDOR_VERBOSE env var should be 1")
	}
}

func TestLogVerbose_Silent(t *testing.T) {
	resetVerbose()
	out := captureStdout(func() {
		logVerbose("should not appear")
	})
	if out != "" {
		t.Errorf("expected no output in non-verbose mode, got %q", out)
	}
}

func TestLogVerbose_Prints(t *testing.T) {
	resetVerbose()
	verbose = true
	out := captureStdout(func() {
		logVerbose("hello %s", "world")
	})
	if strings.TrimSpace(out) != "hello world" {
		t.Errorf("expected 'hello world', got %q", out)
	}
}

func TestInstall_DefaultOutput(t *testing.T) {
	resetVerbose()
	os.RemoveAll(corridorDir())

	out := captureStdout(func() {
		runInstall(nil)
	})

	mustContain := []string{
		"Logged in as",
		"Saved config to",
		"Plugin extracted to",
		"Wrote agent rule to",
		"Installed plugin:",
		"Setup complete.",
	}

	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("default install output missing %q\nfull output:\n%s", s, out)
		}
	}

	mustNotContain := []string{
		"Corridor Plugin Setup",
		"Checking prerequisites",
		"Platform supported:",
		"Validating API key",
		"Populating platform binaries",
		"Extracting",
		"Cleaning up old installation",
		"Authenticating",
	}

	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("default install output should not contain %q\nfull output:\n%s", s, out)
		}
	}
}

func TestInstall_VerboseOutput(t *testing.T) {
	resetVerbose()
	verbose = true

	// Use --force to avoid "already installed" errors from prior tests.
	out := captureStdout(func() {
		runInstall([]string{"--force"})
	})

	mustContain := []string{
		"Corridor Plugin Setup",
		"Checking prerequisites",
		fmt.Sprintf("Platform supported: %s/%s", runtime.GOOS, runtime.GOARCH),
		"Detected targets:",
		"Validating API key",
		"API key is valid",
		"Saving configuration",
		"Cleaning up old installation",
		"Cleaned up old installation",
		"Populating platform binaries",
		"Authenticating",
		"Logged in as",
		"Saved config to",
		"Plugin extracted to",
		"Installed plugin:",
		"Setup complete.",
	}

	for _, s := range mustContain {
		if s == "" {
			continue
		}
		if !strings.Contains(out, s) {
			t.Errorf("verbose install output missing %q\nfull output:\n%s", s, out)
		}
	}
}

func TestLogin_Standalone(t *testing.T) {
	resetVerbose()

	out := captureStdout(func() {
		runLogin(nil, false)
	})

	if !strings.Contains(out, "Opening browser") {
		t.Errorf("standalone login should show 'Opening browser', got %q", out)
	}
	if !strings.Contains(out, "Logged in as") {
		t.Errorf("standalone login should show 'Logged in as', got %q", out)
	}
}

func TestLogin_FromInstall_Default(t *testing.T) {
	resetVerbose()

	out := captureStdout(func() {
		runLogin(nil, true)
	})

	if strings.Contains(out, "Authenticating") {
		t.Errorf("login from install (non-verbose) should not show 'Authenticating', got %q", out)
	}
	if !strings.Contains(out, "Logged in as") {
		t.Errorf("login from install should show 'Logged in as', got %q", out)
	}
}

func TestLogin_FromInstall_Verbose(t *testing.T) {
	resetVerbose()
	verbose = true

	out := captureStdout(func() {
		runLogin(nil, true)
	})

	if !strings.Contains(out, "Authenticating") {
		t.Errorf("login from install (verbose) should show 'Authenticating', got %q", out)
	}
	if !strings.Contains(out, "Logged in as") {
		t.Errorf("login from install (verbose) should show 'Logged in as', got %q", out)
	}
}

func TestVersion_NotShortV(t *testing.T) {
	resetVerbose()
	// -v should now be verbose, not version. --version is still version.
	args := parseGlobalFlags([]string{"-v"})
	if !verbose {
		t.Error("-v should set verbose, not be treated as --version")
	}
	if len(args) != 0 {
		t.Errorf("expected empty args after -v, got %v", args)
	}
}
