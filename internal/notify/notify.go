// Package notify provides macOS notification support via osascript.
// It implements the watcher.Notifier interface.
package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// Notifier sends macOS notifications using osascript.
// It implements the watcher.Notifier interface.
type Notifier struct {
	title   string
	sound   bool
	enabled bool
}

// New creates a new Notifier.
func New(title string, sound bool, enabled bool) *Notifier {
	return &Notifier{
		title:   title,
		sound:   sound,
		enabled: enabled,
	}
}

// Notify sends a notification with the given message.
func (n *Notifier) Notify(message string) error {
	if !n.enabled {
		return nil
	}

	return n.send(n.title, message, "")
}

// NotifyApproval sends a notification for an approval event.
func (n *Notifier) NotifyApproval(paneID, promptType string) error {
	if !n.enabled {
		return nil
	}

	message := fmt.Sprintf("Approved %s in pane %s", promptType, paneID)
	return n.send(n.title, message, "")
}

// NotifyBlocked sends a notification for a blocked command.
func (n *Notifier) NotifyBlocked(paneID string) error {
	if !n.enabled {
		return nil
	}

	message := fmt.Sprintf("Blocked dangerous command in pane %s", paneID)
	return n.send("⚠️ "+n.title, message, "Basso")
}

// send uses osascript to send a macOS notification.
func (n *Notifier) send(title, message, sound string) error {
	// Build the AppleScript
	script := fmt.Sprintf(`display notification %q with title %q`,
		escapeAppleScript(message),
		escapeAppleScript(title),
	)

	if n.sound && sound != "" {
		script += fmt.Sprintf(` sound name %q`, sound)
	} else if n.sound {
		script += ` sound name "default"`
	}

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// escapeAppleScript escapes a string for use in AppleScript.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// IsSupported returns true if notifications are supported on this platform.
func IsSupported() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}
