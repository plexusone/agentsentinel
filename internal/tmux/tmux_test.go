package tmux

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("")
	if c == nil {
		t.Error("NewClient() returned nil")
	}

	c = NewClient("my-session")
	if c == nil {
		t.Error("NewClient() with session returned nil")
	}
}

func TestClient_IsAvailable(t *testing.T) {
	c := NewClient("")
	// tmux should be available on this system
	if !c.IsAvailable() {
		t.Skip("tmux not available, skipping test")
	}
}

func TestClient_IsRunning(t *testing.T) {
	c := NewClient("")
	if !c.IsAvailable() {
		t.Skip("tmux not available, skipping test")
	}
	// Just verify it doesn't panic
	_ = c.IsRunning()
}

func TestClient_ListPanes(t *testing.T) {
	c := NewClient("")
	if !c.IsAvailable() || !c.IsRunning() {
		t.Skip("tmux not available or not running, skipping test")
	}

	panes, err := c.ListPanes()
	if err != nil {
		t.Errorf("ListPanes() error = %v", err)
	}
	if len(panes) == 0 {
		t.Error("ListPanes() returned no panes")
	}
}

func TestClient_ListPanesDetailed(t *testing.T) {
	c := NewClient("")
	if !c.IsAvailable() || !c.IsRunning() {
		t.Skip("tmux not available or not running, skipping test")
	}

	panes, err := c.ListPanesDetailed()
	if err != nil {
		t.Errorf("ListPanesDetailed() error = %v", err)
	}
	if len(panes) == 0 {
		t.Error("ListPanesDetailed() returned no panes")
	}

	// Verify pane structure
	for _, p := range panes {
		if p.ID == "" {
			t.Error("pane ID is empty")
		}
	}
}

func TestClient_CapturePane(t *testing.T) {
	c := NewClient("")
	if !c.IsAvailable() || !c.IsRunning() {
		t.Skip("tmux not available or not running, skipping test")
	}

	panes, err := c.ListPanes()
	if err != nil || len(panes) == 0 {
		t.Skip("no panes available")
	}

	content, err := c.CapturePane(panes[0], 10)
	if err != nil {
		t.Errorf("CapturePane() error = %v", err)
	}
	// Content may be empty, but shouldn't error
	_ = content
}

func TestClient_GetCurrentSession(t *testing.T) {
	c := NewClient("")
	if !c.IsAvailable() || !c.IsRunning() {
		t.Skip("tmux not available or not running, skipping test")
	}

	session, err := c.GetCurrentSession()
	if err != nil {
		// This may fail if not inside tmux
		t.Skipf("GetCurrentSession() error (may not be inside tmux): %v", err)
	}
	if session == "" {
		t.Skip("no current session (may not be inside tmux)")
	}
}
