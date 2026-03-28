package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grokify/mogo/log/slogutil"
	"github.com/lmittmann/tint"
	"github.com/plexusone/agentsentinel/internal/config"
	"github.com/plexusone/agentsentinel/internal/detector"
	"github.com/plexusone/agentsentinel/internal/notify"
	"github.com/plexusone/agentsentinel/internal/stats"
	"github.com/plexusone/agentsentinel/internal/tmux"
	"github.com/plexusone/agentsentinel/internal/watcher"
	"github.com/spf13/cobra"
)

var (
	watchInterval    time.Duration
	watchSession     string
	watchDryRun      bool
	watchBlockDanger bool
	watchLines       int
	watchStats       bool
	watchNotify      bool
	watchLogLevel    string
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch tmux panes and auto-approve tool requests",
	Long: `Watch monitors all tmux panes (or a specific session) for AI coding CLI
tool approval prompts and automatically sends 'y' to approve them.

This works with:
  - Codex CLI
  - Claude Code
  - Gemini CLI
  - AWS Kiro CLI
  - Any CLI that prompts (Y/n) for tool approval

The watcher scans pane contents at a configurable interval and detects
prompts like "Allow? (Y/n)", "Tool request", "Approve tool?", etc.

Safety: By default, dangerous commands (rm -rf, sudo, etc.) are blocked
and require manual approval. Use --no-block-danger to disable this.`,
	Example: `  # Watch all tmux panes
  agentsentinel watch

  # Watch a specific session
  agentsentinel watch --session my-coding-session

  # Watch with faster interval
  agentsentinel watch --interval 500ms

  # Dry run (detect but don't approve)
  agentsentinel watch --dry-run

  # Enable stats and notifications
  agentsentinel watch --stats --notify`,
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 1*time.Second,
		"interval between pane scans")
	watchCmd.Flags().StringVarP(&watchSession, "session", "s", "",
		"tmux session to watch (default: all sessions)")
	watchCmd.Flags().BoolVar(&watchDryRun, "dry-run", false,
		"detect prompts but don't send approval")
	watchCmd.Flags().BoolVar(&watchBlockDanger, "block-danger", true,
		"block dangerous commands from auto-approval")
	watchCmd.Flags().IntVar(&watchLines, "lines", 30,
		"number of lines to capture from each pane")
	watchCmd.Flags().BoolVar(&watchStats, "stats", false,
		"enable statistics tracking")
	watchCmd.Flags().BoolVar(&watchNotify, "notify", false,
		"enable macOS notifications")
	watchCmd.Flags().StringVarP(&watchLogLevel, "level", "l", "info",
		"log level (debug, info, warn, error)")
}

func runWatch(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply config defaults if flags weren't explicitly set
	if !cmd.Flags().Changed("interval") && cfg.Watch.Interval > 0 {
		watchInterval = cfg.Watch.Interval
	}
	if !cmd.Flags().Changed("session") && cfg.Watch.Session != "" {
		watchSession = cfg.Watch.Session
	}
	if !cmd.Flags().Changed("lines") && cfg.Watch.Lines > 0 {
		watchLines = cfg.Watch.Lines
	}
	if !cmd.Flags().Changed("block-danger") {
		watchBlockDanger = cfg.Watch.BlockDanger
	}
	if !cmd.Flags().Changed("stats") {
		watchStats = cfg.Stats.Enabled
	}
	if !cmd.Flags().Changed("notify") {
		watchNotify = cfg.Notifications.Enabled
	}

	logLevel, err := parseLogLevel(watchLogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", watchLogLevel, err)
	}
	// --verbose is a shortcut for --level debug
	if verbose && !cmd.Flags().Changed("level") {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      logLevel,
		TimeFormat: "15:04:05",
	}))

	// Create context with logger for propagation
	ctx := slogutil.ContextWithLogger(context.Background(), logger)

	client := tmux.NewClient(watchSession)

	if !client.IsAvailable() {
		return fmt.Errorf("tmux is not installed or not in PATH")
	}

	if !client.IsRunning() {
		return fmt.Errorf("tmux server is not running")
	}

	// Create detector with custom patterns from config
	det := detector.NewDetector()
	for _, pattern := range cfg.Patterns {
		if err := det.AddPattern(pattern); err != nil {
			logger.Warn("invalid custom pattern", "pattern", pattern, "error", err)
		}
	}
	for _, pattern := range cfg.DangerPatterns {
		if err := det.AddDangerPattern(pattern); err != nil {
			logger.Warn("invalid danger pattern", "pattern", pattern, "error", err)
		}
	}

	// Create notifier
	notifier := notify.New(cfg.Notifications.Title, cfg.Notifications.Sound, watchNotify)

	// Create stats tracker
	st := stats.New()
	if watchStats && cfg.Stats.LogFile != "" {
		if err := st.SetLogFile(cfg.Stats.LogFile); err != nil {
			logger.Warn("failed to open stats log file", "error", err)
		}
	}

	sessionInfo := "all sessions"
	if watchSession != "" {
		sessionInfo = fmt.Sprintf("session '%s'", watchSession)
	}

	logger.Info("starting watcher",
		"session", sessionInfo,
		"interval", watchInterval,
		"dry_run", watchDryRun,
		"block_danger", watchBlockDanger,
		"stats", watchStats,
		"notify", watchNotify,
	)

	// Create watcher with extracted logic
	w := watcher.New(client, det, notifier, st, watcher.Config{
		Lines:       watchLines,
		DryRun:      watchDryRun,
		BlockDanger: watchBlockDanger,
	})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	logger.Info("watching for tool prompts (Ctrl+C to stop)")

	for {
		select {
		case <-ticker.C:
			if err := w.Scan(ctx); err != nil {
				logger.Error("scan error", "error", err)
			}
		case sig := <-sigChan:
			logger.Info("received signal, shutting down", "signal", sig)

			// Print stats on shutdown if enabled
			if watchStats {
				fmt.Println()
				fmt.Println(st.Summary())
			}

			st.Close()
			return nil
		}
	}
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level: %s (valid: debug, info, warn, error)", level)
	}
}
