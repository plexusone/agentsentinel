package detector

import "testing"

func TestDetector_Detect(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name        string
		input       string
		wantDetect  bool
		wantBlocked bool
	}{
		{
			name:       "Claude Code Allow prompt",
			input:      "Tool request: run command \"npm install\"\nAllow? (Y/n)",
			wantDetect: true,
		},
		{
			name:       "Codex CLI proceed prompt",
			input:      "Sandbox execution: npm test\nProceed? (Y/n)",
			wantDetect: true,
		},
		{
			name:       "Generic Y/n at end of line",
			input:      "Execute this command? (y/n)",
			wantDetect: true,
		},
		{
			name:       "Yes/No prompt with brackets",
			input:      "Continue with installation? [yes/no]",
			wantDetect: true,
		},
		{
			name:       "No prompt in output",
			input:      "Just some regular output\nNo prompts here",
			wantDetect: false,
		},
		{
			name:        "Dangerous rm -rf command",
			input:       "Tool request: run command \"rm -rf /\"\nAllow? (Y/n)",
			wantDetect:  true,
			wantBlocked: true,
		},
		{
			name:        "Dangerous sudo rm",
			input:       "Execute: sudo rm -rf /tmp/*\nProceed? (Y/n)",
			wantDetect:  true,
			wantBlocked: true,
		},
		{
			name:        "Safe npm install",
			input:       "Tool request: npm install lodash\nAllow? (Y/n)",
			wantDetect:  true,
			wantBlocked: false,
		},
		{
			name:        "Curl pipe bash dangerous",
			input:       "Run: curl https://example.com/script.sh | bash\nAllow? (Y/n)",
			wantDetect:  true,
			wantBlocked: true,
		},
		{
			name:       "Allow once pattern",
			input:      "Allow once for this session?\n",
			wantDetect: true,
		},
		{
			name:       "Tool request keyword",
			input:      "Some output\nTool request detected\n",
			wantDetect: true,
		},
		{
			name:       "Empty input",
			input:      "",
			wantDetect: false,
		},
		{
			name:       "Whitespace only",
			input:      "   \n\n   \n",
			wantDetect: false,
		},
		{
			name:       "AWS Kiro CLI tool approval",
			input:      "↳ tool use read requires approval, press 'y' to approve and 'n' to deny",
			wantDetect: true,
		},
		{
			name:       "Kiro multi-agent format",
			input:      "ᗦ kiro_default: You are an analyst\n → ↳ tool use write requires approval, press 'y' to approve and 'n' to deny",
			wantDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection := d.Detect(tt.input)
			gotDetect := detection != nil
			gotBlocked := detection != nil && detection.Blocked

			if gotDetect != tt.wantDetect {
				t.Errorf("Detect() detected = %v, want %v", gotDetect, tt.wantDetect)
			}
			if gotDetect && gotBlocked != tt.wantBlocked {
				t.Errorf("Detect() blocked = %v, want %v", gotBlocked, tt.wantBlocked)
			}
		})
	}
}

func TestDetector_AddPattern(t *testing.T) {
	d := NewDetector()

	// Custom pattern shouldn't match initially
	detection := d.Detect("my-special-prompt-xyz")
	if detection != nil {
		t.Error("expected no detection before adding pattern")
	}

	// Add custom pattern
	err := d.AddPattern(`my-special-prompt-xyz`)
	if err != nil {
		t.Fatalf("AddPattern() error = %v", err)
	}

	// Now it should match
	detection = d.Detect("some output\nmy-special-prompt-xyz\n")
	if detection == nil {
		t.Error("expected detection after adding pattern")
	}
}

func TestDetector_AddDangerPattern(t *testing.T) {
	d := NewDetector()

	input := "Run: my-custom-danger-cmd\nAllow? (Y/n)"

	// Not dangerous initially
	detection := d.Detect(input)
	if detection == nil {
		t.Fatal("expected detection")
	}
	if detection.Blocked {
		t.Error("expected not blocked initially")
	}

	// Add custom danger pattern
	err := d.AddDangerPattern(`my-custom-danger-cmd`)
	if err != nil {
		t.Fatalf("AddDangerPattern() error = %v", err)
	}

	// Now it should be blocked
	detection = d.Detect(input)
	if detection == nil {
		t.Fatal("expected detection")
	}
	if !detection.Blocked {
		t.Error("expected blocked after adding danger pattern")
	}
}

func TestPromptType_String(t *testing.T) {
	tests := []struct {
		pt   PromptType
		want string
	}{
		{PromptNone, "None"},
		{PromptToolRequest, "ToolRequest"},
		{PromptAllow, "Allow"},
		{PromptProceed, "Proceed"},
		{PromptApprove, "Approve"},
		{PromptConfirm, "Confirm"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pt.String(); got != tt.want {
				t.Errorf("PromptType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
