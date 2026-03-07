package cmd

import (
	"fmt"

	"github.com/plexusone/agentsentinel/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage AgentSentinel configuration.`,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.ConfigPath())
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("Config file: %s\n\n", config.ConfigPath())
		fmt.Printf("Watch:\n")
		fmt.Printf("  Interval:     %s\n", cfg.Watch.Interval)
		fmt.Printf("  Session:      %s\n", cfg.Watch.Session)
		fmt.Printf("  Lines:        %d\n", cfg.Watch.Lines)
		fmt.Printf("  Block Danger: %v\n", cfg.Watch.BlockDanger)
		fmt.Println()
		fmt.Printf("Patterns:       %d custom\n", len(cfg.Patterns))
		fmt.Printf("Danger Patterns: %d custom\n", len(cfg.DangerPatterns))
		fmt.Println()
		fmt.Printf("Notifications:\n")
		fmt.Printf("  Enabled: %v\n", cfg.Notifications.Enabled)
		fmt.Printf("  Sound:   %v\n", cfg.Notifications.Sound)
		fmt.Println()
		fmt.Printf("Stats:\n")
		fmt.Printf("  Enabled:  %v\n", cfg.Stats.Enabled)
		fmt.Printf("  Log File: %s\n", cfg.Stats.LogFile)

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create example config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		cfg := config.DefaultConfig()
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("Created config file: %s\n", path)
		fmt.Println()
		fmt.Println("Example configuration:")
		fmt.Println(config.Example())

		return nil
	},
}

var configExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Print example configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(config.Example())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configExampleCmd)
}
