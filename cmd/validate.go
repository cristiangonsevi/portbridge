package cmd

import (
	"fmt"
	"os"

	"portbridge/internal/config"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	Long: `Validate the portbridge configuration file for errors and warnings.

Checks each profile for:
  - Required fields (host/ssh_alias, user)
  - Valid port numbers (1-65535)
  - Duplicate tunnel names
  - Missing or non-existent SSH key files
  - Password authentication warnings
  - Other common configuration issues

The command exits with code 0 if no errors are found,
or code 1 if validation errors exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := config.LoadConfig(config.ConfigFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				ui.PrintError(fmt.Sprintf("Configuration file not found: %s", config.ConfigFilePath))
				ui.PrintInfo("Run 'portbridge showconfig' to see the expected path")
				os.Exit(1)
			}
			ui.PrintError(fmt.Sprintf("Failed to load configuration: %v", err))
			os.Exit(1)
		}

		results, configErrors := config.ValidateConfig(*profiles)
		output := config.FormatValidationOutput(*profiles, results, configErrors)

		fmt.Print(output)

		hasErrors := false
		for _, r := range results {
			if len(r.Errors) > 0 {
				hasErrors = true
				break
			}
		}
		if len(configErrors) > 0 {
			hasErrors = true
		}

		if hasErrors {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}
