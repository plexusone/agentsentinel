package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Pane represents a tmux pane.
type Pane struct {
	ID        string
	SessionID string
	WindowID  string
	Index     int
}

// Client provides tmux operations.
type Client struct {
	session string // optional: filter to specific session
}

// NewClient creates a new tmux client.
func NewClient(session string) *Client {
	return &Client{session: session}
}

// IsAvailable checks if tmux is installed and accessible.
func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// IsRunning checks if tmux server is running.
func (c *Client) IsRunning() bool {
	cmd := exec.Command("tmux", "list-sessions")
	err := cmd.Run()
	return err == nil
}

// ListPanes returns all pane IDs in the session (or all sessions if not specified).
func (c *Client) ListPanes() ([]string, error) {
	args := []string{"list-panes", "-a", "-F", "#{pane_id}"}
	if c.session != "" {
		args = []string{"list-panes", "-t", c.session, "-F", "#{pane_id}"}
	}

	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	var panes []string
	for _, line := range lines {
		if len(line) > 0 {
			panes = append(panes, string(line))
		}
	}
	return panes, nil
}

// ListPanesDetailed returns detailed pane information.
func (c *Client) ListPanesDetailed() ([]Pane, error) {
	format := "#{pane_id}:#{session_id}:#{window_id}:#{pane_index}"
	args := []string{"list-panes", "-a", "-F", format}
	if c.session != "" {
		args = []string{"list-panes", "-t", c.session, "-F", format}
	}

	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	var panes []Pane
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(string(line), ":")
		if len(parts) >= 4 {
			idx, _ := strconv.Atoi(parts[3])
			panes = append(panes, Pane{
				ID:        parts[0],
				SessionID: parts[1],
				WindowID:  parts[2],
				Index:     idx,
			})
		}
	}
	return panes, nil
}

// CapturePane returns the visible content of a pane.
// lines specifies how many lines to capture (0 = all visible lines, negative = include scrollback).
func (c *Client) CapturePane(paneID string, lines int) (string, error) {
	args := []string{"capture-pane", "-p", "-t", paneID}
	if lines > 0 {
		// Capture last N lines
		args = append(args, "-S", fmt.Sprintf("-%d", lines))
	}

	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane %s: %w", paneID, err)
	}

	return string(out), nil
}

// SendKeys sends keystrokes to a pane.
func (c *Client) SendKeys(paneID string, keys ...string) error {
	args := []string{"send-keys", "-t", paneID}
	args = append(args, keys...)

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys to pane %s: %w", paneID, err)
	}
	return nil
}

// SendKeysLiteral sends literal text to a pane (no special key interpretation).
func (c *Client) SendKeysLiteral(paneID string, text string) error {
	args := []string{"send-keys", "-t", paneID, "-l", text}

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send literal keys to pane %s: %w", paneID, err)
	}
	return nil
}

// Approve sends 'y' followed by Enter to a pane.
func (c *Client) Approve(paneID string) error {
	return c.SendKeys(paneID, "y", "Enter")
}

// GetCurrentSession returns the current tmux session name.
func (c *Client) GetCurrentSession() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{session_name}").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current session: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
