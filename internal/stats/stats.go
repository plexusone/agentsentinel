// Package stats provides statistics tracking for AgentSentinel approval events.
// It tracks counts by pane and prompt type, maintains recent approval history,
// and optionally logs approvals to a JSON file.
package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/grokify/mogo/log/slogutil"
)

// Approval represents a single approval event recorded by the watcher.
type Approval struct {
	Timestamp time.Time `json:"timestamp"`
	PaneID    string    `json:"pane_id"`
	Type      string    `json:"type"`
	Line      string    `json:"line"`
	Blocked   bool      `json:"blocked"`
}

// Stats tracks approval statistics with thread-safe counters and optional JSON logging.
// It implements the watcher.StatsRecorder interface.
type Stats struct {
	mu sync.Mutex

	// Counters
	TotalApprovals int `json:"total_approvals"`
	TotalBlocked   int `json:"total_blocked"`
	TotalScans     int `json:"total_scans"`

	// Per-pane counts
	ApprovalsByPane map[string]int `json:"approvals_by_pane"`

	// Per-type counts
	ApprovalsByType map[string]int `json:"approvals_by_type"`

	// Recent approvals (last 100)
	RecentApprovals []Approval `json:"recent_approvals"`

	// Start time
	StartTime time.Time `json:"start_time"`

	// Log file (optional)
	logFile *os.File
}

// New creates a new Stats tracker.
func New() *Stats {
	return &Stats{
		ApprovalsByPane: make(map[string]int),
		ApprovalsByType: make(map[string]int),
		RecentApprovals: make([]Approval, 0, 100),
		StartTime:       time.Now(),
	}
}

// SetLogFile sets a file to log approvals to.
func (s *Stats) SetLogFile(path string) error {
	if path == "" {
		return nil
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	s.mu.Lock()
	s.logFile = f
	s.mu.Unlock()

	return nil
}

// Close closes the log file if open.
func (s *Stats) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logFile != nil {
		return s.logFile.Close()
	}
	return nil
}

// RecordScan records a scan event.
func (s *Stats) RecordScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalScans++
}

// RecordApproval records an approval event.
func (s *Stats) RecordApproval(ctx context.Context, paneID, promptType, line string, blocked bool) {
	logger := slogutil.LoggerFromContext(ctx, slog.Default())

	s.mu.Lock()
	defer s.mu.Unlock()

	approval := Approval{
		Timestamp: time.Now(),
		PaneID:    paneID,
		Type:      promptType,
		Line:      line,
		Blocked:   blocked,
	}

	if blocked {
		s.TotalBlocked++
	} else {
		s.TotalApprovals++
		s.ApprovalsByPane[paneID]++
		s.ApprovalsByType[promptType]++
	}

	// Keep last 100 approvals
	if len(s.RecentApprovals) >= 100 {
		s.RecentApprovals = s.RecentApprovals[1:]
	}
	s.RecentApprovals = append(s.RecentApprovals, approval)

	// Log to file if configured
	if s.logFile != nil {
		data, err := json.Marshal(approval)
		if err != nil {
			logger.Warn("failed to marshal approval for log", "error", err)
			return
		}
		if _, err := s.logFile.Write(data); err != nil {
			logger.Warn("failed to write approval to log file", "error", err)
			return
		}
		if _, err := s.logFile.WriteString("\n"); err != nil {
			logger.Warn("failed to write newline to log file", "error", err)
		}
	}
}

// Summary returns a summary of the stats.
func (s *Stats) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	uptime := time.Since(s.StartTime).Round(time.Second)

	return fmt.Sprintf(`Stats Summary
=============
Uptime:          %s
Total Scans:     %d
Total Approvals: %d
Total Blocked:   %d

Approvals by Pane:
%s
Approvals by Type:
%s`,
		uptime,
		s.TotalScans,
		s.TotalApprovals,
		s.TotalBlocked,
		formatMap(s.ApprovalsByPane),
		formatMap(s.ApprovalsByType),
	)
}

// JSON returns stats as JSON.
func (s *Stats) JSON() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.MarshalIndent(s, "", "  ")
}

func formatMap(m map[string]int) string {
	if len(m) == 0 {
		return "  (none)\n"
	}
	var b strings.Builder
	for k, v := range m {
		fmt.Fprintf(&b, "  %s: %d\n", k, v)
	}
	return b.String()
}
