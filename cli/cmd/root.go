package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "corridor",
	Short: "Corridor Security CLI",
	Long:  "A delightful CLI for managing Corridor Security configurations.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
}
