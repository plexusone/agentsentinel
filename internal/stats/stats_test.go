package stats

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()

	if s == nil {
		t.Fatal("expected non-nil Stats")
	}
	if s.TotalApprovals != 0 {
		t.Errorf("expected TotalApprovals 0, got %d", s.TotalApprovals)
	}
	if s.TotalBlocked != 0 {
		t.Errorf("expected TotalBlocked 0, got %d", s.TotalBlocked)
	}
	if s.TotalScans != 0 {
		t.Errorf("expected TotalScans 0, got %d", s.TotalScans)
	}
	if s.ApprovalsByPane == nil {
		t.Error("expected ApprovalsByPane initialized")
	}
	if s.ApprovalsByType == nil {
		t.Error("expected ApprovalsByType initialized")
	}
	if s.RecentApprovals == nil {
		t.Error("expected RecentApprovals initialized")
	}
	if s.StartTime.IsZero() {
		t.Error("expected StartTime set")
	}
}

func TestRecordScan(t *testing.T) {
	s := New()

	s.RecordScan()
	if s.TotalScans != 1 {
		t.Errorf("expected TotalScans 1, got %d", s.TotalScans)
	}

	s.RecordScan()
	s.RecordScan()
	if s.TotalScans != 3 {
		t.Errorf("expected TotalScans 3, got %d", s.TotalScans)
	}
}

func TestRecordApproval_NotBlocked(t *testing.T) {
	s := New()
	ctx := context.Background()

	s.RecordApproval(ctx, "%1", "Allow", "Allow? (Y/n)", false)

	if s.TotalApprovals != 1 {
		t.Errorf("expected TotalApprovals 1, got %d", s.TotalApprovals)
	}
	if s.TotalBlocked != 0 {
		t.Errorf("expected TotalBlocked 0, got %d", s.TotalBlocked)
	}
	if s.ApprovalsByPane["%1"] != 1 {
		t.Errorf("expected ApprovalsByPane[%%1] = 1, got %d", s.ApprovalsByPane["%1"])
	}
	if s.ApprovalsByType["Allow"] != 1 {
		t.Errorf("expected ApprovalsByType[Allow] = 1, got %d", s.ApprovalsByType["Allow"])
	}
	if len(s.RecentApprovals) != 1 {
		t.Errorf("expected 1 recent approval, got %d", len(s.RecentApprovals))
	}

	approval := s.RecentApprovals[0]
	if approval.PaneID != "%1" {
		t.Errorf("expected PaneID '%%1', got %q", approval.PaneID)
	}
	if approval.Type != "Allow" {
		t.Errorf("expected Type 'Allow', got %q", approval.Type)
	}
	if approval.Blocked {
		t.Error("expected Blocked false")
	}
}

func TestRecordApproval_Blocked(t *testing.T) {
	s := New()
	ctx := context.Background()

	s.RecordApproval(ctx, "%2", "Dangerous", "rm -rf /", true)

	if s.TotalApprovals != 0 {
		t.Errorf("expected TotalApprovals 0, got %d", s.TotalApprovals)
	}
	if s.TotalBlocked != 1 {
		t.Errorf("expected TotalBlocked 1, got %d", s.TotalBlocked)
	}
	// Blocked approvals should NOT be counted in per-pane/per-type maps
	if s.ApprovalsByPane["%2"] != 0 {
		t.Errorf("expected ApprovalsByPane[%%2] = 0, got %d", s.ApprovalsByPane["%2"])
	}
	if len(s.RecentApprovals) != 1 {
		t.Errorf("expected 1 recent approval, got %d", len(s.RecentApprovals))
	}
	if !s.RecentApprovals[0].Blocked {
		t.Error("expected Blocked true in RecentApprovals")
	}
}

func TestRecordApproval_CapsAt100(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Add 105 approvals
	for i := 0; i < 105; i++ {
		s.RecordApproval(ctx, "%1", "Allow", "test", false)
	}

	if len(s.RecentApprovals) != 100 {
		t.Errorf("expected RecentApprovals capped at 100, got %d", len(s.RecentApprovals))
	}
	if s.TotalApprovals != 105 {
		t.Errorf("expected TotalApprovals 105, got %d", s.TotalApprovals)
	}
}

func TestSetLogFile_EmptyPath(t *testing.T) {
	s := New()

	err := s.SetLogFile("")
	if err != nil {
		t.Errorf("expected no error for empty path, got %v", err)
	}
	if s.logFile != nil {
		t.Error("expected logFile nil for empty path")
	}
}

func TestSetLogFile_ValidPath(t *testing.T) {
	s := New()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	err := s.SetLogFile(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.logFile == nil {
		t.Fatal("expected logFile to be set")
	}

	// Record an approval and verify it's written
	ctx := context.Background()
	s.RecordApproval(ctx, "%1", "Allow", "test line", false)

	// Close to flush
	if err := s.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	// Read the log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected log file to have content")
	}

	// Verify it's valid JSON
	var approval Approval
	if err := json.Unmarshal(data[:len(data)-1], &approval); err != nil { // -1 for newline
		t.Errorf("failed to parse log entry as JSON: %v", err)
	}
	if approval.PaneID != "%1" {
		t.Errorf("expected PaneID '%%1', got %q", approval.PaneID)
	}
}

func TestSetLogFile_InvalidPath(t *testing.T) {
	s := New()

	err := s.SetLogFile("/nonexistent/directory/file.log")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestClose_NoLogFile(t *testing.T) {
	s := New()

	err := s.Close()
	if err != nil {
		t.Errorf("expected no error when closing without log file, got %v", err)
	}
}

func TestSummary(t *testing.T) {
	s := New()
	ctx := context.Background()

	s.RecordScan()
	s.RecordScan()
	s.RecordApproval(ctx, "%1", "Allow", "test", false)
	s.RecordApproval(ctx, "%2", "Proceed", "test", false)
	s.RecordApproval(ctx, "%1", "Allow", "test", true) // blocked

	summary := s.Summary()

	if !strings.Contains(summary, "Stats Summary") {
		t.Error("expected summary header")
	}
	if !strings.Contains(summary, "Total Scans:     2") {
		t.Error("expected Total Scans: 2")
	}
	if !strings.Contains(summary, "Total Approvals: 2") {
		t.Error("expected Total Approvals: 2")
	}
	if !strings.Contains(summary, "Total Blocked:   1") {
		t.Error("expected Total Blocked: 1")
	}
	if !strings.Contains(summary, "%1: 1") {
		t.Error("expected pane %1 count")
	}
	if !strings.Contains(summary, "Allow: 1") {
		t.Error("expected Allow type count")
	}
}

func TestJSON(t *testing.T) {
	s := New()
	ctx := context.Background()

	s.RecordScan()
	s.RecordApproval(ctx, "%1", "Allow", "test", false)

	data, err := s.JSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("failed to parse JSON: %v", err)
	}

	if parsed["total_scans"].(float64) != 1 {
		t.Errorf("expected total_scans 1, got %v", parsed["total_scans"])
	}
	if parsed["total_approvals"].(float64) != 1 {
		t.Errorf("expected total_approvals 1, got %v", parsed["total_approvals"])
	}
}

func TestMultiplePanes(t *testing.T) {
	s := New()
	ctx := context.Background()

	s.RecordApproval(ctx, "%1", "Allow", "test", false)
	s.RecordApproval(ctx, "%1", "Allow", "test", false)
	s.RecordApproval(ctx, "%2", "Proceed", "test", false)
	s.RecordApproval(ctx, "%3", "Allow", "test", false)

	if s.ApprovalsByPane["%1"] != 2 {
		t.Errorf("expected ApprovalsByPane[%%1] = 2, got %d", s.ApprovalsByPane["%1"])
	}
	if s.ApprovalsByPane["%2"] != 1 {
		t.Errorf("expected ApprovalsByPane[%%2] = 1, got %d", s.ApprovalsByPane["%2"])
	}
	if s.ApprovalsByType["Allow"] != 3 {
		t.Errorf("expected ApprovalsByType[Allow] = 3, got %d", s.ApprovalsByType["Allow"])
	}
	if s.ApprovalsByType["Proceed"] != 1 {
		t.Errorf("expected ApprovalsByType[Proceed] = 1, got %d", s.ApprovalsByType["Proceed"])
	}
}
