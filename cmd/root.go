package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the base command for PortBridge
var RootCmd = &cobra.Command{
	Use:   "portbridge",
	Short: "PortBridge simplifies SSH tunneling with profiles",
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
