// Package watcher implements the core watch loop for auto-approving tool requests.
package watcher

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/grokify/mogo/log/slogutil"
	"github.com/plexusone/agentsentinel/internal/detector"
)

// TmuxClient defines the interface for tmux operations.
// This interface is implemented by internal/tmux.Client.
type TmuxClient interface {
	// ListPanes returns all pane IDs in the monitored session(s).
	ListPanes() ([]string, error)
	// CapturePane captures the last N lines from a pane.
	CapturePane(paneID string, lines int) (string, error)
	// Approve sends a 'y' keystroke to approve a prompt.
	Approve(paneID string) error
	// ApproveMultiple sends multiple 'y' keystrokes with delays for multi-subagent scenarios.
	ApproveMultiple(paneID string, count int, delayMs int) error
}

// Notifier defines the interface for sending notifications.
// This interface is implemented by internal/notify.Notifier.
type Notifier interface {
	// NotifyApproval sends a notification when a prompt is approved.
	NotifyApproval(paneID, promptType string) error
	// NotifyBlocked sends a notification when a dangerous command is blocked.
	NotifyBlocked(paneID string) error
}

// StatsRecorder defines the interface for recording statistics.
// This interface is implemented by internal/stats.Stats.
type StatsRecorder interface {
	// RecordScan increments the scan counter.
	RecordScan()
	// RecordApproval records an approval event with details.
	RecordApproval(ctx context.Context, paneID, promptType, line string, blocked bool)
}

// Config holds watcher configuration.
type Config struct {
	// Lines is the number of lines to capture from each pane.
	Lines int

	// DryRun if true, detects prompts but doesn't send approval.
	DryRun bool

	// BlockDanger if true, blocks dangerous commands from auto-approval.
	BlockDanger bool

	// RecentWindow is the duration to consider a pane as recently approved.
	// Defaults to 5 seconds if zero.
	RecentWindow time.Duration

	// CleanupWindow is the duration after which to clean up old approval records.
	// Defaults to 30 seconds if zero.
	CleanupWindow time.Duration
}

// Watcher monitors tmux panes for approval prompts.
type Watcher struct {
	client   TmuxClient
	detector *detector.Detector
	notifier Notifier
	stats    StatsRecorder
	config   Config

	// Track recently approved panes to avoid duplicate approvals
	recentMu       sync.Mutex
	recentApproved map[string]time.Time
}

// New creates a new Watcher with the given dependencies.
func New(client TmuxClient, det *detector.Detector, notifier Notifier, stats StatsRecorder, cfg Config) *Watcher {
	// Apply defaults
	if cfg.RecentWindow == 0 {
		cfg.RecentWindow = 5 * time.Second
	}
	if cfg.CleanupWindow == 0 {
		cfg.CleanupWindow = 30 * time.Second
	}

	return &Watcher{
		client:         client,
		detector:       det,
		notifier:       notifier,
		stats:          stats,
		config:         cfg,
		recentApproved: make(map[string]time.Time),
	}
}

// Scan performs a single scan of all tmux panes for approval prompts.
// It iterates through all panes, detects prompts using the detector,
// and sends approval keystrokes unless the command is dangerous or
// the watcher is in dry-run mode. Returns an error only if listing
// panes fails; individual pane errors are logged but don't stop the scan.
func (w *Watcher) Scan(ctx context.Context) error {
	logger := slogutil.LoggerFromContext(ctx, slog.Default())

	panes, err := w.client.ListPanes()
	if err != nil {
		return err
	}

	logger.Debug("scanning panes", "count", len(panes))
	w.stats.RecordScan()

	for _, paneID := range panes {
		if w.wasRecentlyApproved(paneID) {
			continue
		}

		content, err := w.client.CapturePane(paneID, w.config.Lines)
		if err != nil {
			logger.Debug("failed to capture pane", "pane", paneID, "error", err)
			continue
		}

		detection := w.detector.Detect(content)
		if detection == nil {
			continue
		}

		detection.PaneID = paneID

		logger.Info("prompt detected",
			"pane", paneID,
			"type", detection.Type.String(),
			"line", Truncate(detection.Line, 60),
			"blocked", detection.Blocked,
		)

		if detection.Blocked && w.config.BlockDanger {
			logger.Warn("dangerous command detected, skipping auto-approval",
				"pane", paneID,
			)
			w.stats.RecordApproval(ctx, paneID, detection.Type.String(), detection.Line, true)
			if err := w.notifier.NotifyBlocked(paneID); err != nil {
				logger.Warn("failed to send notification", "error", err)
			}
			continue
		}

		if w.config.DryRun {
			logger.Info("dry run: would approve", "pane", paneID, "count", detection.Count)
			w.markApproved(paneID)
			continue
		}

		// Handle Kiro-style prompts with multi-subagent TUI navigation
		if detection.Type == detector.PromptApprove && detector.IsKiroPrompt(detection.Line) {
			cycleCount := max(detection.Count, 4) // Kiro typically runs 4 subagents
			logger.Info("kiro multi-subagent detected, cycling through all",
				"pane", paneID,
				"detected", detection.Count,
				"cycles", cycleCount,
			)
			if err := w.client.ApproveMultiple(paneID, cycleCount, 100); err != nil {
				logger.Error("failed to approve multiple", "pane", paneID, "error", err)
				continue
			}
		} else if detection.Count > 1 {
			logger.Info("multi-prompt detected, approving all",
				"pane", paneID,
				"count", detection.Count,
			)
			if err := w.client.ApproveMultiple(paneID, detection.Count, 150); err != nil {
				logger.Error("failed to approve multiple", "pane", paneID, "error", err)
				continue
			}
		} else {
			if err := w.client.Approve(paneID); err != nil {
				logger.Error("failed to approve", "pane", paneID, "error", err)
				continue
			}
		}

		logger.Info("approved", "pane", paneID, "count", detection.Count)
		w.markApproved(paneID)
		w.stats.RecordApproval(ctx, paneID, detection.Type.String(), detection.Line, false)
		if err := w.notifier.NotifyApproval(paneID, detection.Type.String()); err != nil {
			logger.Warn("failed to send notification", "error", err)
		}
	}

	return nil
}

// wasRecentlyApproved checks if a pane was recently approved.
func (w *Watcher) wasRecentlyApproved(paneID string) bool {
	w.recentMu.Lock()
	defer w.recentMu.Unlock()

	if t, ok := w.recentApproved[paneID]; ok {
		if time.Since(t) < w.config.RecentWindow {
			return true
		}
	}
	return false
}

// markApproved marks a pane as recently approved.
func (w *Watcher) markApproved(paneID string) {
	w.recentMu.Lock()
	defer w.recentMu.Unlock()

	w.recentApproved[paneID] = time.Now()

	// Clean up old entries
	for id, t := range w.recentApproved {
		if time.Since(t) > w.config.CleanupWindow {
			delete(w.recentApproved, id)
		}
	}
}

// Truncate truncates a string to the given maximum length, adding "..." suffix
// if truncation occurs. Returns the original string if it fits within maxLen.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
