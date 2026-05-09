package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"portbridge/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of PortBridge",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("portbridge version %s\n", version.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
