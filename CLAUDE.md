# CLAUDE.md

Guidelines for AI agents working with the AgentSentinel codebase.

## Project Overview

**AgentSentinel** is a Go CLI tool that auto-approves tool requests for AI coding assistants (Claude Code, AWS Kiro, Codex CLI, Gemini CLI) running in tmux panes.

**Core functionality:**
1. Monitor tmux panes via `tmux capture-pane`
2. Detect approval prompts using regex patterns
3. Send `y` keystroke via `tmux send-keys` to approve
4. Block dangerous commands (rm -rf, sudo, curl|bash, etc.)

## Architecture

```
main.go                     Entry point
cmd/                        CLI commands (Cobra)
├── root.go                 Root command, --verbose flag
├── watch.go                Watch command setup and main loop
├── status.go               Show tmux status
├── config.go               Configuration management
├── test.go                 Pattern detection tests
├── stats.go                Statistics display
└── version.go              Version info
internal/
├── watcher/                Core watch logic (testable)
│   ├── watcher.go          Watcher struct, Scan(), interfaces
│   └── watcher_test.go     85.7% coverage
├── config/                 YAML configuration
│   ├── config.go           Load/save ~/.agentsentinel.yaml
│   └── config_test.go      73.9% coverage
├── detector/               Prompt detection
│   ├── detector.go         Regex pattern matching, IsKiroPrompt
│   └── detector_test.go    88.9% coverage
├── tmux/                   tmux integration
│   ├── tmux.go             Pane listing, capture, send-keys
│   └── tmux_test.go        53.4% coverage
├── stats/                  Statistics tracking
│   ├── stats.go            Approval counters, JSON logging
│   └── stats_test.go       88.7% coverage
└── notify/                 macOS notifications
    ├── notify.go           osascript wrapper
    └── notify_test.go      50.0% coverage
```

### Key Interfaces

The `internal/watcher` package defines interfaces for dependency injection:

```go
// TmuxClient - tmux operations (implemented by internal/tmux.Client)
type TmuxClient interface {
    ListPanes() ([]string, error)
    CapturePane(paneID string, lines int) (string, error)
    Approve(paneID string) error
    ApproveMultiple(paneID string, count int, delayMs int) error
}

// Notifier - notifications (implemented by internal/notify.Notifier)
type Notifier interface {
    NotifyApproval(paneID, promptType string) error
    NotifyBlocked(paneID string) error
}

// StatsRecorder - statistics (implemented by internal/stats.Stats)
type StatsRecorder interface {
    RecordScan()
    RecordApproval(ctx context.Context, paneID, promptType, line string, blocked bool)
}
```

## Key Patterns

### Context-Based Logging

Logger is propagated via `context.Context`:

```go
import (
    "github.com/grokify/mogo/log/slogutil"
    "github.com/lmittmann/tint"
)

// Create colored logger
logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
    Level:      logLevel,
    TimeFormat: "15:04:05",
}))

// Attach to context
ctx := slogutil.ContextWithLogger(context.Background(), logger)

// Retrieve in functions
func doWork(ctx context.Context) {
    logger := slogutil.LoggerFromContext(ctx, slog.Default())
    logger.Info("working", "key", value)
}
```

### Error Handling

Follow this priority:
1. **Return** - If the function can return an error
2. **Log** - Use `slogutil.LoggerFromContext(ctx, slog.Default())`
3. **Panic** - Only for programming errors / invariant violations

### Adding Detection Patterns

Patterns are in `internal/detector/detector.go`:

```go
// In NewDetector()
patterns: []*regexp.Regexp{
    regexp.MustCompile(`(?i)your-new-pattern`),
    // ...
}
```

Or via config `~/.agentsentinel.yaml`:

```yaml
patterns:
  - "(?i)custom-prompt"
```

### Adding Danger Patterns

```go
dangerPatterns: []*regexp.Regexp{
    regexp.MustCompile(`(?i)dangerous-command`),
}
```

## Development Commands

```bash
# Build
make build

# Test
go test -v ./...

# Test with coverage
go test -cover ./...

# Lint
golangci-lint run

# Install
make install

# Run in dry-run mode
agentsentinel watch --dry-run -v
```

## Configuration

Default config path: `~/.agentsentinel.yaml`

```yaml
watch:
  interval: 1s        # Scan interval
  session: ""         # Filter to specific tmux session
  lines: 30           # Lines to capture per pane
  block_danger: true  # Block dangerous commands

patterns: []          # Custom detection patterns (regex)
danger_patterns: []   # Custom danger patterns (regex)

notifications:
  enabled: false
  sound: true
  title: "AgentSentinel"

stats:
  enabled: false
  log_file: ""        # JSON log file path
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/detector/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Common Tasks

### Add Support for New AI CLI

1. Identify the approval prompt format
2. Add regex pattern to `internal/detector/detector.go` in `NewDetector()`
3. Add test case in `internal/detector/detector_test.go`
4. Update README.md pattern reference table

### Add New Command

1. Create `cmd/newcmd.go`
2. Register in `init()` with `rootCmd.AddCommand(newCmd)`
3. Follow existing command structure (see `cmd/status.go` for simple example)

### Modify Watch Behavior

Core watch logic is in `internal/watcher/watcher.go`:

- `watcher.New()` - Create watcher with dependencies
- `watcher.Scan(ctx)` - Per-tick scan of all panes
- `watcher.wasRecentlyApproved()` - Duplicate prevention (unexported)
- `watcher.markApproved()` - Track approved panes (unexported)

The `cmd/watch.go` handles CLI setup and the main ticker loop:

- `runWatch()` - Load config, create dependencies, start ticker
- Creates `watcher.New(client, detector, notifier, stats, config)`
- Calls `w.Scan(ctx)` on each tick

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML configuration
- `github.com/lmittmann/tint` - Colored slog output
- `github.com/grokify/mogo` - Context-based logging utilities

## File Locations

| What | Where |
|------|-------|
| Config file | `~/.agentsentinel.yaml` |
| Binary (installed) | `$GOPATH/bin/agentsentinel` |
| Main entry | `main.go` |
| Watch command | `cmd/watch.go` |
| Core watcher logic | `internal/watcher/watcher.go` |
| Pattern matching | `internal/detector/detector.go` |
| tmux commands | `internal/tmux/tmux.go` |
| Statistics | `internal/stats/stats.go` |
| Notifications | `internal/notify/notify.go` |
