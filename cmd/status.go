package cmd

import (
	"fmt"

	"github.com/plexusone/agentsentinel/internal/tmux"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show tmux status and available panes",
	Long: `Status displays information about the tmux environment including:
  - Whether tmux is installed
  - Whether tmux server is running
  - List of available panes that can be monitored`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client := tmux.NewClient("")

	fmt.Println("AgentSentinel Status")
	fmt.Println("====================")
	fmt.Println()

	// Check tmux availability
	if !client.IsAvailable() {
		fmt.Println("tmux: NOT INSTALLED")
		fmt.Println()
		fmt.Println("Install tmux with:")
		fmt.Println("  brew install tmux")
		return nil
	}
	fmt.Println("tmux: installed")

	// Check tmux server
	if !client.IsRunning() {
		fmt.Println("tmux server: NOT RUNNING")
		fmt.Println()
		fmt.Println("Start tmux with:")
		fmt.Println("  tmux new-session")
		return nil
	}
	fmt.Println("tmux server: running")

	// Get current session
	session, err := client.GetCurrentSession()
	if err == nil && session != "" {
		fmt.Printf("current session: %s\n", session)
	}

	fmt.Println()

	// List panes
	panes, err := client.ListPanesDetailed()
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	if len(panes) == 0 {
		fmt.Println("No panes found.")
		return nil
	}

	fmt.Printf("Panes (%d total):\n", len(panes))
	fmt.Println()

	for _, pane := range panes {
		fmt.Printf("  %s (session: %s, window: %s, index: %d)\n",
			pane.ID, pane.SessionID, pane.WindowID, pane.Index)
	}

	fmt.Println()
	fmt.Println("Run 'agentsentinel watch' to start monitoring for tool prompts.")

	return nil
}
