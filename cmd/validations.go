package cmd

import (
	"github.com/spf13/cobra"

	"portbridge/internal/config"
)

func getProfileNames() ([]string, error) {
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(*cfg))
	for name := range *cfg {
		names = append(names, name)
	}
	return names, nil
}

func getTunnelNames(profileName string) ([]string, error) {
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	profile, exists := (*cfg)[profileName]
	if !exists {
		return nil, nil
	}
	names := make([]string, 0, len(profile.Tunnels))
	for _, t := range profile.Tunnels {
		names = append(names, t.Name)
	}
	return names, nil
}

func profileNameCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := getProfileNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func tunnelNameCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names, err := getTunnelNames(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
