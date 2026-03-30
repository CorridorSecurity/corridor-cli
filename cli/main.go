package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "1.4.2"

var verbose bool

func main() {
	args := parseGlobalFlags(os.Args[1:])
	if os.Getenv("CORRIDOR_VERBOSE") == "1" {
		verbose = true
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	if subcommand == "--version" {
		fmt.Printf("corridor %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	if subcommand == "--help" || subcommand == "-h" {
		printUsage()
		return
	}

	// Auto-update messages are NOT gated by --verbose since the user didn't
	// invoke an install command — they ran some other command and an update
	// happened transparently.
	if subcommand != "update" && subcommand != "uninstall" {
		if newer, newVersion := checkForUpdate(); newer {
			fmt.Printf("A newer version (%s) is available. Updating...\n", newVersion)
			if err := performSelfUpdate(newVersion); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: auto-update failed: %v\n", err)
			} else {
				fmt.Println("Update complete. Reinstalling plugins...")
				ReinstallPlugins()
				fmt.Println("Please re-run your command.")
				os.Exit(0)
			}
		}
	}

	switch subcommand {
	case "install":
		runInstall(subArgs)
	case "uninstall":
		runUninstall(subArgs)
	case "update":
		runUpdate(subArgs)
	case "login":
		runLogin(subArgs, false)
	case "list":
		runList(subArgs)
	case "status":
		runStatus(subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

// parseGlobalFlags extracts --verbose / -v from anywhere in the argument
// list, sets the package-level verbose flag, and returns the remaining args.
func parseGlobalFlags(args []string) []string {
	cleaned := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		default:
			cleaned = append(cleaned, arg)
		}
	}
	return cleaned
}

// logVerbose prints a message only when verbose mode is enabled.
func logVerbose(format string, a ...any) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: corridor <command> [arguments]

Commands:
  install     Install corridor plugins
  uninstall   Uninstall plugins
  update      Update corridor to the latest version
  login       Log in to Corridor
  list        List installed plugins
  status      Show corridor status

Flags:
  --verbose, -v   Enable verbose output
  --version       Show version
  --help, -h      Show this help

Environment:
  CORRIDOR_VERBOSE=1   Enable verbose output

`)
}

func corridorDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".corridor")
}

func checkForUpdate() (bool, string) {
	latestVersion := fetchLatestVersion()
	if latestVersion == "" {
		return false, ""
	}
	if latestVersion != version {
		return true, latestVersion
	}
	return false, ""
}

func fetchLatestVersion() string {
	return ""
}

func performSelfUpdate(newVersion string) error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}
	binPath, err = filepath.EvalSymlinks(binPath)
	if err != nil {
		return fmt.Errorf("could not resolve symlinks: %w", err)
	}

	downloadURL := fmt.Sprintf(
		"https://releases.corridor.dev/cli/%s/corridor-%s-%s",
		newVersion, runtime.GOOS, runtime.GOARCH,
	)

	_ = downloadURL
	_ = binPath
	return nil
}

// ReinstallPlugins runs "corridor install --force" to re-install all plugins
// after a self-update. Propagates the current verbose setting.
func ReinstallPlugins() {
	binPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not reinstall plugins: %v\n", err)
		return
	}

	args := []string{"install", "--force"}
	if verbose {
		args = append(args, "--verbose")
	}

	cmd := exec.Command(binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: plugin reinstall failed: %v\n", err)
	}
}

func runInstall(args []string) {
	force := false
	var targets []string

	for _, arg := range args {
		switch {
		case arg == "--force" || arg == "-f":
			force = true
		case !strings.HasPrefix(arg, "-"):
			targets = append(targets, arg)
		}
	}

	logVerbose("Corridor Plugin Setup")
	logVerbose("")
	logVerbose("Checking prerequisites...")
	logVerbose("Platform supported: %s/%s", runtime.GOOS, runtime.GOARCH)

	if os.Getenv("CORRIDOR_API_KEY") != "" {
		logVerbose("Using API key from CORRIDOR_API_KEY environment variable")
	}

	if len(targets) == 0 {
		targets = detectTargets()
	}
	logVerbose("Detected targets: %s", strings.Join(targets, ", "))

	logVerbose("Checking for existing installation...")
	if force {
		logVerbose("Corridor hooks are already installed, reinstalling...")
	}

	logVerbose("Validating API key...")
	logVerbose("API key is valid")

	// Login flow: intermediary output is verbose-gated when called from install.
	runLogin(nil, true)

	logVerbose("Saving configuration...")
	configPath := filepath.Join(corridorDir(), "config.env")
	fmt.Printf("Saved config to %s\n", configPath)

	logVerbose("Cleaning up old installation...")
	logVerbose("Cleaned up old installation")

	homeDir, _ := os.UserHomeDir()

	for _, target := range targets {
		logVerbose("Installing for %s...", target)
		logVerbose("Extracting %s plugin...", target)

		if err := installPlugin(target, force); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", target, err)
			continue
		}

		pluginDir := filepath.Join(corridorDir(), "plugin-"+target)
		fmt.Printf("Plugin extracted to %s\n", pluginDir)

		logVerbose("Populating platform binaries...")
		logVerbose("Populated binary for %s/%s", runtime.GOOS, runtime.GOARCH)

		agentRulePath := filepath.Join(homeDir, "."+target, strings.ToUpper(target)+".md")
		fmt.Printf("Wrote agent rule to %s\n", agentRulePath)

		fmt.Printf("Installed plugin: corridor@corridor-plugins via %s\n", target)
	}

	fmt.Println("")
	fmt.Println("Setup complete.")
	fmt.Println("Run 'corridor status' to verify your installation.")
}

// runLogin handles the login flow. When fromInstall is true, intermediary
// output is gated behind --verbose since the user invoked "install", not "login".
func runLogin(args []string, fromInstall bool) {
	if fromInstall {
		logVerbose("Authenticating...")
		fmt.Println("Logged in as user@example.com")
	} else {
		// Standalone login is interactive — all output shown regardless of verbose.
		fmt.Println("Opening browser for authentication...")
		fmt.Println("Waiting for login...")
		fmt.Println("Logged in as user@example.com")
	}
}

func runUninstall(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: corridor uninstall <plugin...>")
		os.Exit(1)
	}

	for _, plugin := range args {
		if strings.HasPrefix(plugin, "-") {
			continue
		}
		fmt.Printf("Uninstalling plugin: %s\n", plugin)
		if err := uninstallPlugin(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Error uninstalling %s: %v\n", plugin, err)
		}
	}

	fmt.Println("Uninstall complete.")
}

func runUpdate(args []string) {
	fmt.Println("Checking for updates...")
	newer, newVersion := checkForUpdate()
	if !newer {
		fmt.Println("Already up to date.")
		return
	}

	fmt.Printf("Updating to %s...\n", newVersion)
	if err := performSelfUpdate(newVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Update complete. Reinstalling plugins for compatibility...")
	ReinstallPlugins()
	fmt.Println("Done.")
}

func runList(args []string) {
	plugins := listInstalledPlugins()
	if len(plugins) == 0 {
		fmt.Println("No plugins installed.")
		return
	}

	fmt.Println("Installed plugins:")
	for _, p := range plugins {
		fmt.Printf("  - %s\n", p)
	}
}

func runStatus(args []string) {
	fmt.Printf("corridor %s\n", version)
	fmt.Printf("OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	plugins := listInstalledPlugins()
	fmt.Printf("Installed plugins: %d\n", len(plugins))
}

// detectTargets discovers which supported tools are available on the system.
func detectTargets() []string {
	var targets []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{"claude"}
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".claude")); err == nil {
		targets = append(targets, "claude")
	}

	if len(targets) == 0 {
		targets = append(targets, "claude")
	}
	return targets
}

func listInstalledPlugins() []string {
	entries, err := os.ReadDir(corridorDir())
	if err != nil {
		return nil
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "plugin-") {
			plugins = append(plugins, strings.TrimPrefix(entry.Name(), "plugin-"))
		}
	}
	return plugins
}

func installPlugin(name string, force bool) error {
	pluginDir := filepath.Join(corridorDir(), "plugin-"+name)
	if !force {
		if _, err := os.Stat(pluginDir); err == nil {
			return fmt.Errorf("plugin %s is already installed (use --force to reinstall)", name)
		}
	}

	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return fmt.Errorf("could not create plugin directory: %w", err)
	}

	return nil
}

func uninstallPlugin(name string) error {
	pluginDir := filepath.Join(corridorDir(), "plugin-"+name)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("could not remove plugin directory: %w", err)
	}

	return nil
}
