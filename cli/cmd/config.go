package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corridorsecurity/corridor-cli/cli/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Corridor configuration",
	Long:  "Interactive configuration management for Corridor Security.",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tui.NewConfigModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}
