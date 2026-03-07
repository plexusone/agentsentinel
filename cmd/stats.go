package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show runtime statistics",
	Long: `Display statistics from a running AgentSentinel watcher.

Note: Stats are tracked per-session. This command shows stats from
the currently running watch process if stats are enabled.

Enable stats in your config file:

  stats:
    enabled: true
    log_file: ~/agentsentinel-approvals.log`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stats are tracked during 'agentsentinel watch' with --stats flag")
		fmt.Println()
		fmt.Println("Enable stats:")
		fmt.Println("  agentsentinel watch --stats")
		fmt.Println()
		fmt.Println("Or in config file (~/.agentsentinel.yaml):")
		fmt.Println("  stats:")
		fmt.Println("    enabled: true")
		fmt.Println("    log_file: ~/agentsentinel-approvals.log")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
