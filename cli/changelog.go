package main

import (
	"fmt"
	"os"
	"strings"
)

type ChangelogEntry struct {
	Version string
	Date    string
	Brief   string
	Details []string
}

// Changelog is ordered newest-first.
var changelog = []ChangelogEntry{
	{
		Version: "1.4.2",
		Date:    "2026-03-30",
		Brief:   "Added --verbose flag, improved install output, checksum verification fix.",
		Details: []string{
			"Add --verbose / -v flag and CORRIDOR_VERBOSE env var for controlling output levels",
			"Gate intermediate install/login output behind --verbose; keep essential output by default",
			"Fix checksum verification in install.sh to fail on mismatch instead of silently continuing",
			"Add --target flag to 'corridor install' for targeting specific agents",
		},
	},
	{
		Version: "1.4.1",
		Date:    "2026-03-24",
		Brief:   "Skip auto-update when running uninstall.",
		Details: []string{
			"Skip auto-update check for 'corridor uninstall' to avoid unnecessary downloads",
		},
	},
	{
		Version: "1.4.0",
		Date:    "2026-03-20",
		Brief:   "Self-update mechanism and plugin reinstall flow.",
		Details: []string{
			"Add auto-update: CLI checks for newer version before running commands",
			"Automatic plugin reinstall after self-update for compatibility",
			"New 'corridor update' command for manual updates",
		},
	},
	{
		Version: "1.3.0",
		Date:    "2026-03-10",
		Brief:   "Config management and status improvements.",
		Details: []string{
			"Add 'corridor config set/get' commands for managing config.env keys",
			"Add 'corridor status --json' for machine-readable status output",
			"Add 'corridor list' command to show installed plugins",
		},
	},
	{
		Version: "1.2.0",
		Date:    "2026-02-28",
		Brief:   "SSO login and multi-target support.",
		Details: []string{
			"Add SSO browser-based login flow via 'corridor login'",
			"Support multiple install targets (auto-detect or specify with --target)",
			"Write agent rules to target-specific directories on install",
		},
	},
	{
		Version: "1.1.0",
		Date:    "2026-02-15",
		Brief:   "Plugin system and uninstall support.",
		Details: []string{
			"Plugin-based architecture for managing tool integrations",
			"Add 'corridor uninstall' command to remove plugins",
			"Add 'corridor install --force' to reinstall existing plugins",
		},
	},
	{
		Version: "1.0.0",
		Date:    "2026-02-01",
		Brief:   "Initial release of Corridor CLI.",
		Details: []string{
			"Install corridor plugins for supported coding agents",
			"Platform detection for linux/darwin on amd64/arm64",
			"Installation script with checksum verification",
		},
	},
}

// FormatBrief returns a single-line summary for the given version.
// If ver is empty, it returns the brief for the latest version.
func FormatBrief(ver string) string {
	entry := findEntry(ver)
	if entry == nil {
		return ""
	}
	return fmt.Sprintf("corridor %s: %s", entry.Version, entry.Brief)
}

// FormatVerbose returns the full changelog for a given version.
// If ver is empty, it returns the full changelog for all versions.
func FormatVerbose(ver string) string {
	if ver != "" {
		entry := findEntry(ver)
		if entry == nil {
			return fmt.Sprintf("No changelog found for version %s.", ver)
		}
		return formatSingleEntry(entry)
	}

	var sb strings.Builder
	sb.WriteString("Corridor CLI Changelog\n")
	sb.WriteString("======================\n\n")
	for i, entry := range changelog {
		sb.WriteString(formatSingleEntry(&entry))
		if i < len(changelog)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// UpdateBrief returns a short message suitable for displaying after an
// auto-update. Shows the brief for the version we just updated to.
func UpdateBrief(newVersion string) string {
	entry := findEntry(newVersion)
	if entry == nil {
		return fmt.Sprintf("Updated to corridor %s.", newVersion)
	}
	return fmt.Sprintf("Updated to corridor %s — %s\nRun 'corridor changelog' for full details.", entry.Version, entry.Brief)
}

func findEntry(ver string) *ChangelogEntry {
	if ver == "" && len(changelog) > 0 {
		return &changelog[0]
	}
	for i := range changelog {
		if changelog[i].Version == ver {
			return &changelog[i]
		}
	}
	return nil
}

func formatSingleEntry(entry *ChangelogEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%s)\n", entry.Version, entry.Date))
	sb.WriteString(fmt.Sprintf("  %s\n\n", entry.Brief))
	for _, detail := range entry.Details {
		sb.WriteString(fmt.Sprintf("  - %s\n", detail))
	}
	sb.WriteString("\n")
	return sb.String()
}

func runChangelog(args []string) {
	brief := false
	targetVersion := ""

	for _, arg := range args {
		switch {
		case arg == "--brief" || arg == "-b":
			brief = true
		case !strings.HasPrefix(arg, "-"):
			targetVersion = arg
		}
	}

	if brief {
		ver := targetVersion
		if ver == "" {
			ver = version
		}
		out := FormatBrief(ver)
		if out == "" {
			fmt.Fprintf(os.Stderr, "No changelog found for version %s.\n", ver)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	fmt.Print(FormatVerbose(targetVersion))
}
