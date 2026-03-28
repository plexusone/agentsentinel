// Package detector provides regex-based detection of AI CLI tool approval prompts.
// It identifies prompts from Claude Code, AWS Kiro, Codex CLI, Gemini CLI, and
// other tools that request user confirmation before executing actions.
//
// The detector also identifies dangerous commands (rm -rf, sudo, curl|bash, etc.)
// that should be blocked from auto-approval.
package detector

import (
	"regexp"
	"strings"
)

// PromptType identifies the type of tool prompt detected.
type PromptType int

const (
	PromptNone PromptType = iota
	PromptToolRequest
	PromptAllow
	PromptProceed
	PromptApprove
	PromptConfirm
)

// Detection represents a detected tool prompt with metadata.
type Detection struct {
	// Type categorizes the prompt (Allow, Approve, Proceed, etc.)
	Type PromptType
	// Line is the actual text that matched a detection pattern.
	Line string
	// PaneID is the tmux pane identifier (set by the watcher, not detector).
	PaneID string
	// Blocked is true if the surrounding text contains dangerous commands.
	Blocked bool
	// Count is the number of pending prompts (for multi-subagent scenarios like Kiro).
	Count int
}

// Detector scans text for tool approval prompts using configurable regex patterns.
// It maintains separate pattern sets for prompt detection and dangerous command detection.
type Detector struct {
	patterns       []*regexp.Regexp
	dangerPatterns []*regexp.Regexp
}

// NewDetector creates a new prompt detector with built-in patterns for
// Claude Code, AWS Kiro, Codex CLI, Gemini CLI, and generic Y/n prompts.
// Custom patterns can be added via AddPattern and AddDangerPattern.
func NewDetector() *Detector {
	return &Detector{
		patterns: []*regexp.Regexp{
			// Generic Y/n prompts
			regexp.MustCompile(`(?i)\(y/n\)\s*$`),
			regexp.MustCompile(`(?i)\[y/n\]\s*$`),
			regexp.MustCompile(`(?i)\(yes/no\)\s*$`),
			regexp.MustCompile(`(?i)\[yes/no\]\s*$`),

			// Tool request patterns
			regexp.MustCompile(`(?i)allow\s*\?\s*\(y/n\)`),
			regexp.MustCompile(`(?i)allow\s+tool`),
			regexp.MustCompile(`(?i)tool\s+request`),
			regexp.MustCompile(`(?i)approve\s+tool`),
			regexp.MustCompile(`(?i)proceed\s*\?`),
			regexp.MustCompile(`(?i)continue\s*\?\s*\(y/n\)`),
			regexp.MustCompile(`(?i)execute\s*\?\s*\(y/n\)`),
			regexp.MustCompile(`(?i)run\s+command\s*\?`),

			// Claude Code specific
			regexp.MustCompile(`(?i)allow\s+once`),
			regexp.MustCompile(`(?i)allow\s+always`),

			// Codex CLI specific
			regexp.MustCompile(`(?i)sandbox\s+execution`),

			// AWS Kiro CLI specific
			regexp.MustCompile(`(?i)tool\s+use\s+\w+\s+requires\s+approval`),
			regexp.MustCompile(`(?i)press\s+'y'\s+to\s+approve`),

			// Generic permission prompts
			regexp.MustCompile(`(?i)permission\s+required`),
			regexp.MustCompile(`(?i)confirm\s+action`),
		},
		dangerPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)rm\s+-rf`),
			regexp.MustCompile(`(?i)rm\s+-r\s+/`),
			regexp.MustCompile(`(?i)sudo\s+rm`),
			regexp.MustCompile(`(?i)mkfs`),
			regexp.MustCompile(`(?i)dd\s+if=`),
			regexp.MustCompile(`(?i):\(\)\s*\{\s*:\|:\s*&\s*\}`), // fork bomb
			regexp.MustCompile(`(?i)>\s*/dev/sd`),
			regexp.MustCompile(`(?i)chmod\s+-R\s+777\s+/`),
			regexp.MustCompile(`(?i)chown\s+-R.*\s+/`),
			regexp.MustCompile(`(?i)curl.*\|\s*sh`),
			regexp.MustCompile(`(?i)curl.*\|\s*bash`),
			regexp.MustCompile(`(?i)wget.*\|\s*sh`),
			regexp.MustCompile(`(?i)wget.*\|\s*bash`),
		},
	}
}

// kiroMultiPromptPattern matches Kiro's multi-subagent approval prompts
var kiroMultiPromptPattern = regexp.MustCompile(`(?i)tool\s+use\s+\w+\s+requires\s+approval.*press\s+'y'\s+to\s+approve`)

// Detect scans the given text for tool prompts, checking the last 30 lines
// for matches against configured patterns. Returns nil if no prompt is found.
// If a prompt is detected, the returned Detection includes the prompt type,
// matched line, whether it's blocked (dangerous), and count of pending prompts.
func (d *Detector) Detect(text string) *Detection {
	lines := strings.Split(text, "\n")

	// Check last 30 lines for prompts (most recent content)
	start := 0
	if len(lines) > 30 {
		start = len(lines) - 30
	}

	// First, count Kiro multi-subagent prompts
	kiroCount := d.countKiroPrompts(text)

	for i := len(lines) - 1; i >= start; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Skip lines that look like log output or structured data
		if d.isLogLine(line) {
			continue
		}

		for _, pattern := range d.patterns {
			if pattern.MatchString(line) {
				detection := &Detection{
					Type:    d.classifyPrompt(line),
					Line:    line,
					Blocked: d.isDangerous(text),
					Count:   max(1, kiroCount), // At least 1, or the Kiro count
				}
				return detection
			}
		}
	}

	return nil
}

// countKiroPrompts counts the number of Kiro-style multi-subagent approval prompts.
func (d *Detector) countKiroPrompts(text string) int {
	matches := kiroMultiPromptPattern.FindAllString(text, -1)
	return len(matches)
}

// isLogLine returns true if the line looks like log output (not a real prompt).
func (d *Detector) isLogLine(line string) bool {
	// Skip lines with common log patterns
	logIndicators := []string{
		"level=",
		"msg=",
		"time=",
		"[INFO]",
		"[DEBUG]",
		"[WARN]",
		"[ERROR]",
		"INFO:",
		"DEBUG:",
		"WARN:",
		"ERROR:",
	}

	for _, indicator := range logIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}

	return false
}

// classifyPrompt determines the type of prompt.
func (d *Detector) classifyPrompt(line string) PromptType {
	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "tool request"):
		return PromptToolRequest
	case strings.Contains(lower, "allow"):
		return PromptAllow
	case strings.Contains(lower, "proceed"):
		return PromptProceed
	case strings.Contains(lower, "approve"):
		return PromptApprove
	case strings.Contains(lower, "confirm"):
		return PromptConfirm
	default:
		return PromptAllow
	}
}

// isDangerous checks if the text contains dangerous commands.
func (d *Detector) isDangerous(text string) bool {
	for _, pattern := range d.dangerPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// AddPattern adds a custom regex pattern for prompt detection.
// Returns an error if the pattern is invalid.
func (d *Detector) AddPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	d.patterns = append(d.patterns, re)
	return nil
}

// AddDangerPattern adds a custom regex pattern for dangerous command detection.
// Prompts containing text matching danger patterns will be blocked from auto-approval.
// Returns an error if the pattern is invalid.
func (d *Detector) AddDangerPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	d.dangerPatterns = append(d.dangerPatterns, re)
	return nil
}

// String returns a string representation of the prompt type.
func (t PromptType) String() string {
	switch t {
	case PromptToolRequest:
		return "ToolRequest"
	case PromptAllow:
		return "Allow"
	case PromptProceed:
		return "Proceed"
	case PromptApprove:
		return "Approve"
	case PromptConfirm:
		return "Confirm"
	default:
		return "None"
	}
}

// IsKiroPrompt checks if the line matches AWS Kiro CLI's approval prompt format.
// Kiro prompts contain "tool use", "requires approval", and "press 'y' to approve".
func IsKiroPrompt(line string) bool {
	return strings.Contains(line, "tool use") &&
		strings.Contains(line, "requires approval") &&
		strings.Contains(line, "press 'y' to approve")
}
