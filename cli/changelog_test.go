package main

import (
	"strings"
	"testing"
)

func TestChangelogNotEmpty(t *testing.T) {
	if len(changelog) == 0 {
		t.Fatal("changelog must contain at least one entry")
	}
}

func TestChangelogNewestFirst(t *testing.T) {
	if changelog[0].Version != version {
		t.Errorf("first changelog entry should match current version %s, got %s", version, changelog[0].Version)
	}
}

func TestFormatBrief_CurrentVersion(t *testing.T) {
	out := FormatBrief(version)
	if out == "" {
		t.Fatalf("FormatBrief(%q) returned empty string", version)
	}
	if !strings.HasPrefix(out, "corridor "+version+": ") {
		t.Errorf("expected prefix 'corridor %s: ', got %q", version, out)
	}
}

func TestFormatBrief_LatestWhenEmpty(t *testing.T) {
	out := FormatBrief("")
	if out == "" {
		t.Fatal("FormatBrief(\"\") returned empty string; should return latest")
	}
	if !strings.Contains(out, changelog[0].Version) {
		t.Errorf("expected latest version %s in output, got %q", changelog[0].Version, out)
	}
}

func TestFormatBrief_UnknownVersion(t *testing.T) {
	out := FormatBrief("99.99.99")
	if out != "" {
		t.Errorf("expected empty string for unknown version, got %q", out)
	}
}

func TestFormatVerbose_SingleVersion(t *testing.T) {
	out := FormatVerbose(version)
	if !strings.Contains(out, "## "+version) {
		t.Errorf("verbose output should contain version header, got:\n%s", out)
	}
	for _, detail := range changelog[0].Details {
		if !strings.Contains(out, detail) {
			t.Errorf("verbose output missing detail %q", detail)
		}
	}
}

func TestFormatVerbose_AllVersions(t *testing.T) {
	out := FormatVerbose("")
	if !strings.Contains(out, "Corridor CLI Changelog") {
		t.Errorf("full changelog should contain title header")
	}
	for _, entry := range changelog {
		if !strings.Contains(out, "## "+entry.Version) {
			t.Errorf("full changelog missing version %s", entry.Version)
		}
	}
}

func TestFormatVerbose_UnknownVersion(t *testing.T) {
	out := FormatVerbose("99.99.99")
	if !strings.Contains(out, "No changelog found") {
		t.Errorf("expected 'No changelog found' for unknown version, got %q", out)
	}
}

func TestUpdateBrief_KnownVersion(t *testing.T) {
	out := UpdateBrief(version)
	if !strings.Contains(out, "Updated to corridor "+version) {
		t.Errorf("expected 'Updated to corridor %s' in output, got %q", version, out)
	}
	if !strings.Contains(out, changelog[0].Brief) {
		t.Errorf("expected brief text in output, got %q", out)
	}
	if !strings.Contains(out, "corridor changelog") {
		t.Errorf("should mention 'corridor changelog' command, got %q", out)
	}
}

func TestUpdateBrief_UnknownVersion(t *testing.T) {
	out := UpdateBrief("99.99.99")
	if !strings.Contains(out, "Updated to corridor 99.99.99") {
		t.Errorf("expected fallback message, got %q", out)
	}
}

func TestRunChangelog_Brief(t *testing.T) {
	resetVerbose()
	out := captureStdout(func() {
		runChangelog([]string{"--brief"})
	})
	if !strings.Contains(out, "corridor "+version) {
		t.Errorf("brief changelog should contain current version, got %q", out)
	}
	if strings.Contains(out, "##") {
		t.Errorf("brief changelog should not contain section headers, got %q", out)
	}
}

func TestRunChangelog_BriefShortFlag(t *testing.T) {
	resetVerbose()
	out := captureStdout(func() {
		runChangelog([]string{"-b"})
	})
	if !strings.Contains(out, "corridor "+version) {
		t.Errorf("-b flag should produce brief changelog, got %q", out)
	}
}

func TestRunChangelog_Verbose(t *testing.T) {
	resetVerbose()
	out := captureStdout(func() {
		runChangelog(nil)
	})
	if !strings.Contains(out, "Corridor CLI Changelog") {
		t.Errorf("default changelog should be verbose with title, got %q", out)
	}
	for _, entry := range changelog {
		if !strings.Contains(out, entry.Version) {
			t.Errorf("verbose changelog missing version %s", entry.Version)
		}
	}
}

func TestRunChangelog_SpecificVersion(t *testing.T) {
	resetVerbose()
	target := changelog[len(changelog)-1].Version
	out := captureStdout(func() {
		runChangelog([]string{target})
	})
	if !strings.Contains(out, "## "+target) {
		t.Errorf("expected version header for %s, got %q", target, out)
	}
}

func TestRunChangelog_BriefSpecificVersion(t *testing.T) {
	resetVerbose()
	target := changelog[len(changelog)-1].Version
	out := captureStdout(func() {
		runChangelog([]string{"--brief", target})
	})
	expected := "corridor " + target + ": "
	if !strings.HasPrefix(strings.TrimSpace(out), expected) {
		t.Errorf("expected brief for %s, got %q", target, out)
	}
}

func TestChangelogEntries_HaveRequiredFields(t *testing.T) {
	for i, entry := range changelog {
		if entry.Version == "" {
			t.Errorf("changelog[%d] has empty Version", i)
		}
		if entry.Date == "" {
			t.Errorf("changelog[%d] (%s) has empty Date", i, entry.Version)
		}
		if entry.Brief == "" {
			t.Errorf("changelog[%d] (%s) has empty Brief", i, entry.Version)
		}
		if len(entry.Details) == 0 {
			t.Errorf("changelog[%d] (%s) has no Details", i, entry.Version)
		}
	}
}
