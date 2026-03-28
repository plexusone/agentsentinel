package watcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/plexusone/agentsentinel/internal/detector"
)

// Mock implementations for testing

type mockTmuxClient struct {
	panes        []string
	paneContents map[string]string
	listErr      error
	captureErr   error
	approveErr   error
	approvals    []string
	multiApprove []struct {
		paneID  string
		count   int
		delayMs int
	}
}

func (m *mockTmuxClient) ListPanes() ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.panes, nil
}

func (m *mockTmuxClient) CapturePane(paneID string, lines int) (string, error) {
	if m.captureErr != nil {
		return "", m.captureErr
	}
	if content, ok := m.paneContents[paneID]; ok {
		return content, nil
	}
	return "", nil
}

func (m *mockTmuxClient) Approve(paneID string) error {
	if m.approveErr != nil {
		return m.approveErr
	}
	m.approvals = append(m.approvals, paneID)
	return nil
}

func (m *mockTmuxClient) ApproveMultiple(paneID string, count int, delayMs int) error {
	if m.approveErr != nil {
		return m.approveErr
	}
	m.multiApprove = append(m.multiApprove, struct {
		paneID  string
		count   int
		delayMs int
	}{paneID, count, delayMs})
	return nil
}

type mockNotifier struct {
	approvals []struct {
		paneID     string
		promptType string
	}
	blocked    []string
	approveErr error
	blockErr   error
}

func (m *mockNotifier) NotifyApproval(paneID, promptType string) error {
	if m.approveErr != nil {
		return m.approveErr
	}
	m.approvals = append(m.approvals, struct {
		paneID     string
		promptType string
	}{paneID, promptType})
	return nil
}

func (m *mockNotifier) NotifyBlocked(paneID string) error {
	if m.blockErr != nil {
		return m.blockErr
	}
	m.blocked = append(m.blocked, paneID)
	return nil
}

type mockStats struct {
	scans     int
	approvals []struct {
		paneID     string
		promptType string
		line       string
		blocked    bool
	}
}

func (m *mockStats) RecordScan() {
	m.scans++
}

func (m *mockStats) RecordApproval(ctx context.Context, paneID, promptType, line string, blocked bool) {
	m.approvals = append(m.approvals, struct {
		paneID     string
		promptType string
		line       string
		blocked    bool
	}{paneID, promptType, line, blocked})
}

// Tests

func TestNew(t *testing.T) {
	client := &mockTmuxClient{}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:       30,
		DryRun:      false,
		BlockDanger: true,
	})

	if w == nil {
		t.Fatal("expected non-nil Watcher")
	}
	if w.config.RecentWindow != 5*time.Second {
		t.Errorf("expected default RecentWindow 5s, got %v", w.config.RecentWindow)
	}
	if w.config.CleanupWindow != 30*time.Second {
		t.Errorf("expected default CleanupWindow 30s, got %v", w.config.CleanupWindow)
	}
}

func TestScan_NoPanes(t *testing.T) {
	client := &mockTmuxClient{panes: []string{}}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stats.scans != 1 {
		t.Errorf("expected 1 scan recorded, got %d", stats.scans)
	}
}

func TestScan_ListPanesError(t *testing.T) {
	client := &mockTmuxClient{listErr: errors.New("tmux error")}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err == nil {
		t.Error("expected error")
	}
}

func TestScan_NoPromptDetected(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "normal output\nno prompts here\n",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 0 {
		t.Errorf("expected no approvals, got %d", len(client.approvals))
	}
}

func TestScan_PromptDetected_Approved(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "Allow? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:       30,
		BlockDanger: true,
	})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 1 {
		t.Errorf("expected 1 approval, got %d", len(client.approvals))
	}
	if client.approvals[0] != "%1" {
		t.Errorf("expected approval for %%1, got %s", client.approvals[0])
	}
	if len(stats.approvals) != 1 || stats.approvals[0].blocked {
		t.Error("expected non-blocked approval recorded")
	}
}

func TestScan_DryRun(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "Allow? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:  30,
		DryRun: true,
	})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 0 {
		t.Errorf("expected no approvals in dry run, got %d", len(client.approvals))
	}
}

func TestScan_DangerousCommand_Blocked(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "rm -rf /\nAllow? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:       30,
		BlockDanger: true,
	})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 0 {
		t.Errorf("expected no approvals for dangerous command, got %d", len(client.approvals))
	}
	if len(stats.approvals) != 1 || !stats.approvals[0].blocked {
		t.Error("expected blocked approval recorded")
	}
	if len(notifier.blocked) != 1 {
		t.Errorf("expected 1 blocked notification, got %d", len(notifier.blocked))
	}
}

func TestScan_DangerousCommand_NotBlocked_WhenDisabled(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "rm -rf /\nAllow? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:       30,
		BlockDanger: false, // Danger blocking disabled
	})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 1 {
		t.Errorf("expected 1 approval when danger blocking disabled, got %d", len(client.approvals))
	}
}

func TestScan_RecentlyApproved_Skipped(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "Allow? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:        30,
		RecentWindow: 5 * time.Second,
	})
	ctx := context.Background()

	// First scan - should approve
	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 1 {
		t.Errorf("expected 1 approval on first scan, got %d", len(client.approvals))
	}

	// Second scan immediately - should skip
	err = w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 1 {
		t.Errorf("expected still 1 approval after second scan, got %d", len(client.approvals))
	}
}

func TestScan_MultiplePanes(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1", "%2", "%3"},
		paneContents: map[string]string{
			"%1": "Allow? (Y/n)",
			"%2": "normal output",
			"%3": "Proceed? (Y/n)",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.approvals) != 2 {
		t.Errorf("expected 2 approvals, got %d", len(client.approvals))
	}
}

func TestScan_KiroPrompt_MultiApprove(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "tool use read requires approval, press 'y' to approve",
		},
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(client.multiApprove) != 1 {
		t.Errorf("expected 1 multi-approve call, got %d", len(client.multiApprove))
	}
	if client.multiApprove[0].count < 4 {
		t.Errorf("expected at least 4 cycles for Kiro, got %d", client.multiApprove[0].count)
	}
}

func TestScan_ApproveError(t *testing.T) {
	client := &mockTmuxClient{
		panes: []string{"%1"},
		paneContents: map[string]string{
			"%1": "Allow? (Y/n)",
		},
		approveErr: errors.New("approve failed"),
	}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{Lines: 30})
	ctx := context.Background()

	// Should not return error, just log it
	err := w.Scan(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Stats should not record approval since it failed
	if len(stats.approvals) != 0 {
		t.Errorf("expected no approvals recorded on error, got %d", len(stats.approvals))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"empty string", "", 10, ""},
		{"truncate longer", "abcdefgh", 6, "abc..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Truncate(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
			}
		})
	}
}

func TestWasRecentlyApproved(t *testing.T) {
	client := &mockTmuxClient{}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:        30,
		RecentWindow: 100 * time.Millisecond,
	})

	// Initially not approved
	if w.wasRecentlyApproved("%1") {
		t.Error("expected pane not to be recently approved initially")
	}

	// Mark as approved
	w.markApproved("%1")

	// Should be recently approved
	if !w.wasRecentlyApproved("%1") {
		t.Error("expected pane to be recently approved after marking")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should no longer be recently approved
	if w.wasRecentlyApproved("%1") {
		t.Error("expected pane to not be recently approved after window expired")
	}
}

func TestMarkApproved_Cleanup(t *testing.T) {
	client := &mockTmuxClient{}
	det := detector.NewDetector()
	notifier := &mockNotifier{}
	stats := &mockStats{}

	w := New(client, det, notifier, stats, Config{
		Lines:         30,
		CleanupWindow: 50 * time.Millisecond,
	})

	// Mark multiple panes
	w.markApproved("%1")
	w.markApproved("%2")

	// Both should be tracked
	w.recentMu.Lock()
	if len(w.recentApproved) != 2 {
		t.Errorf("expected 2 tracked panes, got %d", len(w.recentApproved))
	}
	w.recentMu.Unlock()

	// Wait for cleanup window
	time.Sleep(100 * time.Millisecond)

	// Mark a new pane - this triggers cleanup
	w.markApproved("%3")

	// Old entries should be cleaned up
	w.recentMu.Lock()
	if len(w.recentApproved) != 1 {
		t.Errorf("expected 1 tracked pane after cleanup, got %d", len(w.recentApproved))
	}
	if _, ok := w.recentApproved["%3"]; !ok {
		t.Error("expected %3 to be tracked")
	}
	w.recentMu.Unlock()
}
