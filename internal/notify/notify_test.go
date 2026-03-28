package notify

import (
	"testing"
)

func TestNew(t *testing.T) {
	n := New("TestTitle", true, true)

	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
	if n.title != "TestTitle" {
		t.Errorf("expected title 'TestTitle', got %q", n.title)
	}
	if !n.sound {
		t.Error("expected sound true")
	}
	if !n.enabled {
		t.Error("expected enabled true")
	}
}

func TestNew_Disabled(t *testing.T) {
	n := New("Title", false, false)

	if n.enabled {
		t.Error("expected enabled false")
	}
	if n.sound {
		t.Error("expected sound false")
	}
}

func TestNotify_Disabled(t *testing.T) {
	n := New("Title", true, false) // disabled

	err := n.Notify("test message")
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestNotifyApproval_Disabled(t *testing.T) {
	n := New("Title", true, false) // disabled

	err := n.NotifyApproval("%1", "Allow")
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestNotifyBlocked_Disabled(t *testing.T) {
	n := New("Title", true, false) // disabled

	err := n.NotifyBlocked("%1")
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escaping needed",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "escape backslash",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "escape quotes",
			input:    `say "hello"`,
			expected: `say \"hello\"`,
		},
		{
			name:     "escape both",
			input:    `say "hello" and \n`,
			expected: `say \"hello\" and \\n`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple quotes",
			input:    `"a" "b" "c"`,
			expected: `\"a\" \"b\" \"c\"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeAppleScript(tc.input)
			if result != tc.expected {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	// This just tests that the function doesn't panic
	// The actual result depends on the platform
	supported := IsSupported()
	t.Logf("IsSupported() = %v", supported)
}

// Note: We don't test actual notification sending because:
// 1. It requires osascript (macOS only)
// 2. It would actually display notifications during tests
// 3. The disabled path is the important one to test

func TestNotifier_EnabledButNoOsascript(t *testing.T) {
	// Skip this test if osascript is available (i.e., on macOS)
	// because we don't want to actually send notifications
	if IsSupported() {
		t.Skip("skipping on macOS where osascript is available")
	}

	n := New("Title", true, true) // enabled

	// On non-macOS, this should fail because osascript doesn't exist
	err := n.Notify("test")
	if err == nil {
		t.Error("expected error when osascript not available")
	}
}
