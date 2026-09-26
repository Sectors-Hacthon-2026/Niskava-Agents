package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell autocompletion scripts for Niskava Agent.
Supports Bash, Zsh, Fish, and PowerShell.

Example (Bash):
  $ source <(niskava completion bash)

Example (Zsh):
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ niskava completion zsh > "${fpath[1]}/_niskava"

Example (PowerShell):
  PS> niskava completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Shell completion generation does not require database connection
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return RootCmd.GenBashCompletion(out)
		case "zsh":
			return RootCmd.GenZshCompletion(out)
		case "fish":
			return RootCmd.GenFishCompletion(out, true)
		case "powershell":
			return RootCmd.GenPowerShellCompletionWithDesc(out)
		default:
			return fmt.Errorf("unsupported shell type: %s", args[0])
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
