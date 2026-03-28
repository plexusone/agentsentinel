package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Watch.Interval != 1*time.Second {
		t.Errorf("expected interval 1s, got %v", cfg.Watch.Interval)
	}
	if cfg.Watch.Lines != 30 {
		t.Errorf("expected lines 30, got %d", cfg.Watch.Lines)
	}
	if !cfg.Watch.BlockDanger {
		t.Error("expected block_danger true")
	}
	if cfg.Notifications.Enabled {
		t.Error("expected notifications disabled by default")
	}
	if !cfg.Notifications.Sound {
		t.Error("expected notifications sound true by default")
	}
	if cfg.Notifications.Title != "AgentSentinel" {
		t.Errorf("expected title 'AgentSentinel', got %q", cfg.Notifications.Title)
	}
	if cfg.Stats.Enabled {
		t.Error("expected stats disabled by default")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("expected non-empty config path")
	}
	if !filepath.IsAbs(path) && path != ".agentsentinel.yaml" {
		t.Errorf("expected absolute path or fallback, got %q", path)
	}
}

func TestLoadFrom_NonExistent(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("expected no error for non-existent file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}
	// Should return defaults
	if cfg.Watch.Interval != 1*time.Second {
		t.Errorf("expected default interval, got %v", cfg.Watch.Interval)
	}
}

func TestLoadFrom_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
watch:
  interval: 2s
  session: "test-session"
  lines: 50
  block_danger: false
patterns:
  - "custom-pattern"
notifications:
  enabled: true
  sound: false
  title: "TestTitle"
stats:
  enabled: true
  log_file: "/tmp/test.log"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Watch.Interval != 2*time.Second {
		t.Errorf("expected interval 2s, got %v", cfg.Watch.Interval)
	}
	if cfg.Watch.Session != "test-session" {
		t.Errorf("expected session 'test-session', got %q", cfg.Watch.Session)
	}
	if cfg.Watch.Lines != 50 {
		t.Errorf("expected lines 50, got %d", cfg.Watch.Lines)
	}
	if cfg.Watch.BlockDanger {
		t.Error("expected block_danger false")
	}
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "custom-pattern" {
		t.Errorf("expected patterns ['custom-pattern'], got %v", cfg.Patterns)
	}
	if !cfg.Notifications.Enabled {
		t.Error("expected notifications enabled")
	}
	if cfg.Notifications.Sound {
		t.Error("expected notifications sound false")
	}
	if cfg.Notifications.Title != "TestTitle" {
		t.Errorf("expected title 'TestTitle', got %q", cfg.Notifications.Title)
	}
	if !cfg.Stats.Enabled {
		t.Error("expected stats enabled")
	}
	if cfg.Stats.LogFile != "/tmp/test.log" {
		t.Errorf("expected log_file '/tmp/test.log', got %q", cfg.Stats.LogFile)
	}
}

func TestLoadFrom_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
watch:
  interval: [invalid
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadFrom(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveToAndLoadFrom_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	original := &Config{
		Watch: WatchConfig{
			Interval:    3 * time.Second,
			Session:     "my-session",
			Lines:       40,
			BlockDanger: true,
		},
		Patterns:       []string{"pattern1", "pattern2"},
		DangerPatterns: []string{"danger1"},
		Notifications: NotifyConfig{
			Enabled: true,
			Sound:   true,
			Title:   "MyTitle",
		},
		Stats: StatsConfig{
			Enabled: true,
			LogFile: "/var/log/test.log",
		},
	}

	if err := original.SaveTo(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Watch.Interval != original.Watch.Interval {
		t.Errorf("interval mismatch: got %v, want %v", loaded.Watch.Interval, original.Watch.Interval)
	}
	if loaded.Watch.Session != original.Watch.Session {
		t.Errorf("session mismatch: got %q, want %q", loaded.Watch.Session, original.Watch.Session)
	}
	if loaded.Watch.Lines != original.Watch.Lines {
		t.Errorf("lines mismatch: got %d, want %d", loaded.Watch.Lines, original.Watch.Lines)
	}
	if len(loaded.Patterns) != len(original.Patterns) {
		t.Errorf("patterns length mismatch: got %d, want %d", len(loaded.Patterns), len(original.Patterns))
	}
	if loaded.Notifications.Title != original.Notifications.Title {
		t.Errorf("title mismatch: got %q, want %q", loaded.Notifications.Title, original.Notifications.Title)
	}
}

func TestExample(t *testing.T) {
	example := Example()
	if example == "" {
		t.Error("expected non-empty example")
	}
	if len(example) < 100 {
		t.Error("expected substantial example content")
	}
}
