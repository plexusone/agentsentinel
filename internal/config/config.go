package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	// Watch settings
	Watch WatchConfig `yaml:"watch"`

	// Custom patterns to detect
	Patterns []string `yaml:"patterns,omitempty"`

	// Custom danger patterns to block
	DangerPatterns []string `yaml:"danger_patterns,omitempty"`

	// Notifications settings
	Notifications NotifyConfig `yaml:"notifications"`

	// Stats settings
	Stats StatsConfig `yaml:"stats"`
}

// WatchConfig holds watch command settings.
type WatchConfig struct {
	Interval    time.Duration `yaml:"interval"`
	Session     string        `yaml:"session,omitempty"`
	Lines       int           `yaml:"lines"`
	BlockDanger bool          `yaml:"block_danger"`
}

// NotifyConfig holds notification settings.
type NotifyConfig struct {
	Enabled bool   `yaml:"enabled"`
	Sound   bool   `yaml:"sound"`
	Title   string `yaml:"title,omitempty"`
}

// StatsConfig holds stats settings.
type StatsConfig struct {
	Enabled bool   `yaml:"enabled"`
	LogFile string `yaml:"log_file,omitempty"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Watch: WatchConfig{
			Interval:    1 * time.Second,
			Lines:       30,
			BlockDanger: true,
		},
		Notifications: NotifyConfig{
			Enabled: false,
			Sound:   true,
			Title:   "AgentSentinel",
		},
		Stats: StatsConfig{
			Enabled: false,
		},
	}
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".agentsentinel.yaml"
	}
	return filepath.Join(home, ".agentsentinel.yaml")
}

// Load loads configuration from file, falling back to defaults.
func Load() (*Config, error) {
	return LoadFrom(ConfigPath())
}

// LoadFrom loads configuration from the specified path.
func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if no config file
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to the default path.
func (c *Config) Save() error {
	return c.SaveTo(ConfigPath())
}

// SaveTo saves the configuration to the specified path.
func (c *Config) SaveTo(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Example returns an example configuration as YAML string.
func Example() string {
	return `# AgentSentinel Configuration
# Place this file at ~/.agentsentinel.yaml

watch:
  # Scan interval (e.g., 1s, 500ms)
  interval: 1s
  # Specific tmux session to watch (empty = all sessions)
  session: ""
  # Number of lines to capture from each pane
  lines: 30
  # Block dangerous commands from auto-approval
  block_danger: true

# Custom patterns to detect (in addition to built-in patterns)
patterns:
  # - "my-custom-prompt"
  # - "approve this action"

# Custom dangerous command patterns to block
danger_patterns:
  # - "drop database"
  # - "format disk"

notifications:
  # Enable macOS notifications
  enabled: false
  # Play sound with notification
  sound: true
  # Notification title
  title: "AgentSentinel"

stats:
  # Enable stats tracking
  enabled: false
  # Log file for approval history
  log_file: ""
`
}
