package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// listCmd lists all profiles in the configuration
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig("~/.config/.portbridge/portbridge.yaml")
		if err != nil {
			ui.PrintError("Failed to load configuration: " + err.Error())
			return
		}

		ui.PrintInfo("Available profiles:")
		for profileName := range *config {
			ui.PrintInfo("- " + profileName)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
