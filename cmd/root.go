package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"

	rootCmd = &cobra.Command{
		Use:   "agentsentinel",
		Short: "Auto-approve tool requests for AI coding CLIs",
		Long: `AgentSentinel monitors tmux panes for AI coding CLI tool requests
(Codex CLI, Claude Code, Gemini CLI, AWS Kiro CLI) and automatically
responds with 'Y' to approve them.

It works by scanning tmux pane contents for tool approval prompts
and sending keystrokes to the appropriate pane.`,
	}
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
}
