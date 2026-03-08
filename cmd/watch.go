package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/plexusone/agentsentinel/internal/config"
	"github.com/plexusone/agentsentinel/internal/detector"
	"github.com/plexusone/agentsentinel/internal/notify"
	"github.com/plexusone/agentsentinel/internal/stats"
	"github.com/plexusone/agentsentinel/internal/tmux"
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
}

type watcher struct {
	client      *tmux.Client
	detector    *detector.Detector
	logger      *slog.Logger
	notifier    *notify.Notifier
	stats       *stats.Stats
	dryRun      bool
	blockDanger bool

	// Track recently approved panes to avoid duplicate approvals
	recentMu       sync.Mutex
	recentApproved map[string]time.Time
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

	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

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
	st.SetLogger(logger)
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

	w := &watcher{
		client:         client,
		detector:       det,
		logger:         logger,
		notifier:       notifier,
		stats:          st,
		dryRun:         watchDryRun,
		blockDanger:    watchBlockDanger,
		recentApproved: make(map[string]time.Time),
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	logger.Info("watching for tool prompts (Ctrl+C to stop)")

	for {
		select {
		case <-ticker.C:
			if err := w.scan(); err != nil {
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

func (w *watcher) scan() error {
	panes, err := w.client.ListPanes()
	if err != nil {
		return err
	}

	w.logger.Debug("scanning panes", "count", len(panes))
	w.stats.RecordScan()

	for _, paneID := range panes {
		if w.wasRecentlyApproved(paneID) {
			continue
		}

		content, err := w.client.CapturePane(paneID, watchLines)
		if err != nil {
			w.logger.Debug("failed to capture pane", "pane", paneID, "error", err)
			continue
		}

		detection := w.detector.Detect(content)
		if detection == nil {
			continue
		}

		detection.PaneID = paneID

		w.logger.Info("prompt detected",
			"pane", paneID,
			"type", detection.Type.String(),
			"line", truncate(detection.Line, 60),
			"blocked", detection.Blocked,
		)

		if detection.Blocked && w.blockDanger {
			w.logger.Warn("dangerous command detected, skipping auto-approval",
				"pane", paneID,
			)
			w.stats.RecordApproval(paneID, detection.Type.String(), detection.Line, true)
			if err := w.notifier.NotifyBlocked(paneID); err != nil {
				w.logger.Warn("failed to send notification", "error", err)
			}
			continue
		}

		if w.dryRun {
			w.logger.Info("dry run: would approve", "pane", paneID, "count", detection.Count)
			w.markApproved(paneID)
			continue
		}

		// Handle Kiro-style prompts with multi-subagent TUI navigation
		// Even if count=1, use ApproveMultiple to cycle through in case cursor
		// isn't on the pending item. Use minimum of 4 cycles for Kiro prompts.
		if detection.Type == detector.PromptApprove && isKiroPrompt(detection.Line) {
			cycleCount := detection.Count
			if cycleCount < 4 {
				cycleCount = 4 // Kiro typically runs 4 subagents
			}
			w.logger.Info("kiro multi-subagent detected, cycling through all",
				"pane", paneID,
				"detected", detection.Count,
				"cycles", cycleCount,
			)
			if err := w.client.ApproveMultiple(paneID, cycleCount, 100); err != nil {
				w.logger.Error("failed to approve multiple", "pane", paneID, "error", err)
				continue
			}
		} else if detection.Count > 1 {
			w.logger.Info("multi-prompt detected, approving all",
				"pane", paneID,
				"count", detection.Count,
			)
			if err := w.client.ApproveMultiple(paneID, detection.Count, 150); err != nil {
				w.logger.Error("failed to approve multiple", "pane", paneID, "error", err)
				continue
			}
		} else {
			if err := w.client.Approve(paneID); err != nil {
				w.logger.Error("failed to approve", "pane", paneID, "error", err)
				continue
			}
		}

		w.logger.Info("approved", "pane", paneID, "count", detection.Count)
		w.markApproved(paneID)
		w.stats.RecordApproval(paneID, detection.Type.String(), detection.Line, false)
		if err := w.notifier.NotifyApproval(paneID, detection.Type.String()); err != nil {
			w.logger.Warn("failed to send notification", "error", err)
		}
	}

	return nil
}

func (w *watcher) wasRecentlyApproved(paneID string) bool {
	w.recentMu.Lock()
	defer w.recentMu.Unlock()

	if t, ok := w.recentApproved[paneID]; ok {
		// Don't re-approve within 5 seconds
		if time.Since(t) < 5*time.Second {
			return true
		}
	}
	return false
}

func (w *watcher) markApproved(paneID string) {
	w.recentMu.Lock()
	defer w.recentMu.Unlock()

	w.recentApproved[paneID] = time.Now()

	// Clean up old entries
	for id, t := range w.recentApproved {
		if time.Since(t) > 30*time.Second {
			delete(w.recentApproved, id)
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// isKiroPrompt checks if the line matches Kiro's approval prompt format.
func isKiroPrompt(line string) bool {
	return strings.Contains(line, "tool use") &&
		strings.Contains(line, "requires approval") &&
		strings.Contains(line, "press 'y' to approve")
}
