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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	if subcommand == "--version" || subcommand == "-v" {
		fmt.Printf("corridor %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	if subcommand == "--help" || subcommand == "-h" {
		printUsage()
		return
	}

	// Auto-update: check for a newer version before running any command,
	// except "update" (handled explicitly) and "uninstall" (must not
	// reinstall plugins the user is about to remove).
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
		runInstall(os.Args[2:])
	case "uninstall":
		runUninstall(os.Args[2:])
	case "update":
		runUpdate(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "status":
		runStatus(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: corridor <command> [arguments]

Commands:
  install     Install plugins
  uninstall   Uninstall plugins
  update      Update corridor to the latest version
  list        List installed plugins
  status      Show corridor status

Flags:
  --version, -v   Show version
  --help, -h      Show this help

`)
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

// ReinstallPlugins runs "corridor install --force" to ensure all plugins
// are compatible with the new CLI version after a self-update.
func ReinstallPlugins() {
	binPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not reinstall plugins: %v\n", err)
		return
	}

	cmd := exec.Command(binPath, "install", "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: plugin reinstall failed: %v\n", err)
	}
}

func runInstall(args []string) {
	force := false
	plugins := []string{}

	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else if !strings.HasPrefix(arg, "-") {
			plugins = append(plugins, arg)
		}
	}

	if len(plugins) == 0 {
		plugins = discoverPlugins()
	}

	for _, plugin := range plugins {
		if force {
			fmt.Printf("Force-installing plugin: %s\n", plugin)
		} else {
			fmt.Printf("Installing plugin: %s\n", plugin)
		}
		if err := installPlugin(plugin, force); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", plugin, err)
		}
	}

	fmt.Println("Install complete.")
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
	plugins := discoverPlugins()
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

	plugins := discoverPlugins()
	fmt.Printf("Installed plugins: %d\n", len(plugins))
}

func discoverPlugins() []string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	pluginsDir := filepath.Join(configDir, "corridor", "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			plugins = append(plugins, entry.Name())
		}
	}
	return plugins
}

func installPlugin(name string, force bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not determine config directory: %w", err)
	}

	pluginDir := filepath.Join(configDir, "corridor", "plugins", name)
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
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not determine config directory: %w", err)
	}

	pluginDir := filepath.Join(configDir, "corridor", "plugins", name)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("could not remove plugin directory: %w", err)
	}

	return nil
}
