package cmd

import (
	"bufio"
	"fmt"
	"os"
	"portbridge/internal/config"

	"github.com/spf13/cobra"
)

// showConfigCmd displays the current configuration file path and state file path
var showConfigCmd = &cobra.Command{
	Use:   "showconfig",
	Short: "Show the current configuration and default paths",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration File Path:", config.ConfigFilePath)
		fmt.Println("State File Path:", config.StateFilePath)

		fmt.Println("\nFirst 5 lines of the configuration file:")
		file, err := os.Open(config.ConfigFilePath)
		if err != nil {
			fmt.Println("Error reading configuration file:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for i := 0; i < 5 && scanner.Scan(); i++ {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning configuration file:", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(showConfigCmd)
}
