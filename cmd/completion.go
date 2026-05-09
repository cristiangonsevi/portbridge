package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for PortBridge.

To load completions:

Bash:

  $ source <(portbridge completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ portbridge completion bash > /etc/bash_completion.d/portbridge
  # macOS:
  $ portbridge completion bash > /usr/local/etc/bash_completion.d/portbridge

Zsh:

  # If shell completion is not already enabled in your zsh environment:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ portbridge completion zsh > "${fpath[1]}/_portbridge"

  # You will need to start a new shell for the setup to take effect.

Fish:

  $ portbridge completion fish | source

  # To load completions for each session, execute once:
  $ portbridge completion fish > ~/.config/fish/completions/portbridge.fish

PowerShell:

  $ portbridge completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, execute once:
  $ portbridge completion powershell > portbridge.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			if err := RootCmd.GenBashCompletion(os.Stdout); err != nil {
				os.Stderr.WriteString("Error generating bash completion: " + err.Error() + "\n")
			}
		case "zsh":
			if err := RootCmd.GenZshCompletion(os.Stdout); err != nil {
				os.Stderr.WriteString("Error generating zsh completion: " + err.Error() + "\n")
			}
		case "fish":
			if err := RootCmd.GenFishCompletion(os.Stdout, true); err != nil {
				os.Stderr.WriteString("Error generating fish completion: " + err.Error() + "\n")
			}
		case "powershell":
			if err := RootCmd.GenPowerShellCompletionWithDesc(os.Stdout); err != nil {
				os.Stderr.WriteString("Error generating powershell completion: " + err.Error() + "\n")
			}
		}
	},
}
