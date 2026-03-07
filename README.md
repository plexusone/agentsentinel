# AgentSentinel

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

Auto-approve tool requests for AI coding CLIs running in tmux.

AgentSentinel monitors tmux panes for tool approval prompts from AI coding assistants and automatically responds with `y` to approve them. Designed for macOS with iTerm2 and tmux.

## Supported AI CLIs

- **AWS Kiro CLI** - `tool use read requires approval, press 'y' to approve`
- **Claude Code** - `Allow? (Y/n)`
- **Codex CLI** - `Proceed? (Y/n)`
- **Gemini CLI** - `(Y/n)` prompts

## Requirements

- macOS
- tmux 3.0+
- Go 1.21+ (for building)

## Installation

### Option 1: Install to PATH (recommended)

```bash
# Install to ~/go/bin (make sure ~/go/bin is in your PATH)
go install github.com/plexusone/agentsentinel@latest

# Verify installation
agentsentinel version
```

### Option 2: Build locally

```bash
git clone https://github.com/plexusone/agentsentinel.git
cd agentsentinel
go build -o agentsentinel .

# Run from current directory
./agentsentinel version
```

## Quick Start

### Step 1: Start tmux first

```bash
# In iTerm2, start tmux
tmux new-session -s agents
```

### Step 2: Run your AI CLI inside tmux

```bash
# You're now inside tmux, start your AI CLI
kiro
# or: claude
# or: codex
```

### Step 3: Run AgentSentinel (3 options)

**Option A: In another tmux pane (recommended)**

```bash
# Split the tmux window
# Press: Ctrl+b "  (horizontal split)
# or:    Ctrl+b %  (vertical split)

# In the new pane, run:
agentsentinel watch -v
```

**Option B: In a separate iTerm2 tab (outside tmux)**

```bash
# Open new iTerm2 tab (Cmd+T)
# Don't start tmux, just run directly:
agentsentinel watch -v
```

**Option C: Background tmux session**

```bash
# From anywhere, create a detached session:
tmux new-session -d -s sentinel 'agentsentinel watch -v'
```

### Step 4: Watch it work

```
┌─────────────────────────────────────┐
│ tmux pane 1: kiro                   │
│                                     │
│ > tool use read requires approval,  │
│   press 'y' to approve              │  ← AgentSentinel detects this
│                                     │
├─────────────────────────────────────┤
│ tmux pane 2: agentsentinel watch -v │
│                                     │
│ INFO prompt detected pane=%1        │
│ INFO approved pane=%1               │  ← Sends 'y' to pane 1
└─────────────────────────────────────┘
```

### Key Point

**AgentSentinel monitors tmux from the outside** - it uses `tmux` commands to read pane contents and send keystrokes. It doesn't run "inside" your AI CLI. It runs separately and watches all tmux panes.

### Quick Test

```bash
# Check AgentSentinel can see your tmux panes
agentsentinel status

# Dry run to see what it would approve (without actually approving)
agentsentinel watch --dry-run -v
```

## Advanced Usage

### Where to Run AgentSentinel

AgentSentinel communicates with tmux via commands (`tmux list-panes`, `tmux send-keys`), so it works from anywhere:

| Location | Pros | Cons |
|----------|------|------|
| **Tmux pane** | See logs alongside agents | Uses a pane |
| **Separate iTerm2 tab** | Doesn't use tmux space | Switch tabs to see logs |
| **Background session** | Out of the way | Need to attach to see logs |

### Multi-Agent Layout (3 panes)

Running multiple AI agents with AgentSentinel in a dedicated pane:

```
┌────────────────────────────────────────────────────────┐
│ iTerm2                                                 │
│ ┌─────────────────────┬─────────────────────┐          │
│ │ tmux pane 1         │ tmux pane 2         │          │
│ │ $ kiro              │ $ claude            │          │
│ │                     │                     │          │
│ │ Agent working...    │ Agent working...    │          │
│ │ tool use requires   │ Allow? (Y/n)        │          │
│ │ approval, press 'y' │        ↑            │          │
│ │        ↑            │        │            │          │
│ │        │            │        │            │          │
│ │        └────────────┴────────┘            │          │
│ │                     │                     │          │
│ ├─────────────────────┴─────────────────────┤          │
│ │ tmux pane 3                               │          │
│ │ $ agentsentinel watch -v                  │          │
│ │ INFO starting watcher session="all"       │          │
│ │ INFO prompt detected pane=%1 type=Approve │          │
│ │ INFO approved pane=%1                     │          │
│ └───────────────────────────────────────────┘          │
└────────────────────────────────────────────────────────┘
```

### Multi-Agent Workflows

AgentSentinel handles multiple concurrent agents. When running 4+ subagents (like Kiro's parallel execution), it monitors all panes and approves whichever one has a pending prompt.

```bash
# Watch all tmux sessions and panes
agentsentinel watch

# Watch a specific tmux session
agentsentinel watch --session my-agents

# Faster scanning for responsive approval (default: 1s)
agentsentinel watch --interval 500ms
```

### Background with Logging

```bash
# Run in background, log to file
agentsentinel watch -v > ~/agentsentinel.log 2>&1 &

# Check logs
tail -f ~/agentsentinel.log

# Stop it
pkill agentsentinel
```

## Commands

### watch

Monitor tmux panes and auto-approve tool requests.

```bash
agentsentinel watch [flags]

Flags:
  -i, --interval duration   Scan interval (default 1s)
  -s, --session string      Watch specific tmux session only
      --dry-run             Detect but don't approve
      --block-danger        Block dangerous commands (default true)
      --lines int           Lines to capture per pane (default 30)
      --stats               Enable statistics tracking
      --notify              Enable macOS notifications
  -v, --verbose             Verbose output
```

### status

Show tmux environment and available panes.

```bash
agentsentinel status
```

Example output:

```
AgentSentinel Status
====================

tmux: installed
tmux server: running
current session: dev

Panes (4 total):

  %1 (session: $0, window: @0, index: 0)
  %2 (session: $0, window: @0, index: 1)
  %3 (session: $0, window: @1, index: 0)
  %4 (session: $1, window: @2, index: 0)

Run 'agentsentinel watch' to start monitoring for tool prompts.
```

### test

Verify prompt detection is working.

```bash
agentsentinel test
```

### version

```bash
agentsentinel version
```

### config

Manage configuration file.

```bash
# Show config file path
agentsentinel config path

# Show current configuration
agentsentinel config show

# Create example config file
agentsentinel config init

# Print example configuration
agentsentinel config example
```

### completion

Generate shell completion scripts.

```bash
# Bash
agentsentinel completion bash > /usr/local/etc/bash_completion.d/agentsentinel

# Zsh
agentsentinel completion zsh > "${fpath[1]}/_agentsentinel"

# Fish
agentsentinel completion fish > ~/.config/fish/completions/agentsentinel.fish
```

## Configuration

AgentSentinel can be configured via `~/.agentsentinel.yaml`:

```yaml
watch:
  interval: 1s
  session: ""
  lines: 30
  block_danger: true

# Custom patterns to detect
patterns:
  - "my-custom-prompt"

# Custom dangerous command patterns
danger_patterns:
  - "drop database"

notifications:
  enabled: true
  sound: true
  title: "AgentSentinel"

stats:
  enabled: true
  log_file: ~/agentsentinel-approvals.log
```

Run `agentsentinel config example` for a full example with comments.

## Safety Features

### Dangerous Command Blocking

By default, AgentSentinel blocks auto-approval for dangerous commands:

- `rm -rf`
- `sudo rm`
- `mkfs`
- `dd if=`
- `curl ... | bash`
- `wget ... | sh`
- `chmod -R 777 /`

When detected, you'll see:

```
WARN dangerous command detected, skipping auto-approval pane=%3
```

To disable (not recommended):

```bash
agentsentinel watch --block-danger=false
```

### Dry Run Mode

Always test with `--dry-run` first to see what would be approved:

```bash
agentsentinel watch --dry-run -v
```

### Duplicate Prevention

AgentSentinel tracks recently approved panes to prevent sending multiple approvals to the same prompt.

## Notifications (macOS)

Enable macOS notifications to see when approvals happen:

```bash
agentsentinel watch --notify
```

Or in config:

```yaml
notifications:
  enabled: true
  sound: true
```

## Statistics

Track approval statistics during a session:

```bash
agentsentinel watch --stats
```

Stats are printed on shutdown (Ctrl+C). Enable logging to file:

```yaml
stats:
  enabled: true
  log_file: ~/agentsentinel-approvals.log
```

## How It Works

1. **Pane Discovery** - Lists all tmux panes via `tmux list-panes`
2. **Content Capture** - Reads each pane's visible content via `tmux capture-pane`
3. **Pattern Matching** - Scans for approval prompts using regex patterns
4. **Keystroke Injection** - Sends `y` + Enter via `tmux send-keys`

This approach is more reliable than arrow-key navigation because it directly targets the correct pane regardless of UI layout.

## Detected Patterns

| Pattern | Example |
|---------|---------|
| `(Y/n)` | `Allow? (Y/n)` |
| `[yes/no]` | `Continue? [yes/no]` |
| `tool use ... requires approval` | Kiro CLI |
| `press 'y' to approve` | Kiro CLI |
| `Allow once` | Claude Code |
| `Proceed?` | Codex CLI |
| `Tool request` | Generic |

## Troubleshooting

### "tmux is not installed"

```bash
brew install tmux
```

### "tmux server is not running"

Start a tmux session first:

```bash
tmux new-session
```

### Prompts not being detected

1. Run with verbose logging: `agentsentinel watch -v`
2. Check if the prompt appears in the last 30 lines of the pane
3. Run `agentsentinel test` to verify detection patterns
4. The prompt may use a different format - open an issue with the exact prompt text

### Approvals not working

1. Verify the pane ID with `agentsentinel status`
2. Test manually: `tmux send-keys -t %1 y Enter`
3. Some CLIs may require focus - check if the CLI is waiting for input

## Development

```bash
# Run tests
go test -v ./...

# Lint
golangci-lint run

# Build
go build -o agentsentinel .
```

## License

MIT

 [go-ci-svg]: https://github.com/plexusone/agentsentinel/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/agentsentinel/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/agentsentinel/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/agentsentinel/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/agentsentinel/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/agentsentinel/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/agentsentinel
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/agentsentinel
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/agentsentinel
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/agentsentinel
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fagentsentinel
 [loc-svg]: https://tokei.rs/b1/github/plexusone/agentsentinel
 [repo-url]: https://github.com/plexusone/agentsentinel
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/agentsentinel/blob/master/LICENSE
