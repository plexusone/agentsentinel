# Release Notes - v0.1.0

**Release Date:** 2026-03-07

## Overview

AgentSentinel v0.1.0 is the initial release of the tmux-based auto-approval tool for AI coding CLIs. It monitors tmux panes for tool approval prompts from AI assistants and automatically responds with approval, enabling hands-free operation of multiple concurrent AI agents.

## Highlights

- **Multi-CLI Support** - Works with AWS Kiro CLI, Claude Code, Codex CLI, and Gemini CLI
- **Kiro Multi-Subagent Support** - Automatically cycles through and approves all concurrent Kiro subagents in a single pane using TUI navigation (y+j)
- **Safety First** - Built-in dangerous command blocking prevents auto-approval of destructive operations
- **Multi-Agent Ready** - Monitor all tmux panes simultaneously for parallel AI agent workflows

## Supported AI CLIs

| CLI | Prompt Pattern |
|-----|----------------|
| AWS Kiro | `tool use ... requires approval, press 'y' to approve` |
| Claude Code | `Allow? (Y/n)`, `Allow once` |
| Codex CLI | `Proceed? (Y/n)` |
| Gemini CLI | `(Y/n)` prompts |

## Commands

- `watch` - Monitor tmux panes and auto-approve tool requests
- `status` - Display tmux environment and available panes
- `test` - Verify prompt detection patterns
- `config` - Manage configuration (path, show, init, example)
- `stats` - Track approval statistics
- `version` - Display version information

## Key Features

### Auto-Approval
Monitor all tmux panes and automatically send `y` + Enter when approval prompts are detected:
```bash
agentsentinel watch -v
```

### Dangerous Command Blocking
By default, auto-approval is blocked for dangerous commands:
- `rm -rf`, `sudo rm`
- `mkfs`, `dd if=`
- `curl ... | bash`, `wget ... | sh`
- `chmod -R 777 /`

### Dry Run Mode
Test detection without actually sending approvals:
```bash
agentsentinel watch --dry-run -v
```

### Session Filtering
Watch only a specific tmux session:
```bash
agentsentinel watch --session my-agents
```

### macOS Notifications
Get notified when approvals happen:
```bash
agentsentinel watch --notify
```

## Configuration

AgentSentinel can be configured via `~/.agentsentinel.yaml`:

```yaml
watch:
  interval: 1s
  session: ""
  lines: 30
  block_danger: true

notifications:
  enabled: true
  sound: true
```

Run `agentsentinel config example` for a full configuration template.

## Requirements

- macOS
- tmux 3.0+
- Go 1.21+ (for building from source)

## Installation

```bash
# Install from source
go install github.com/plexusone/agentsentinel@latest

# Verify installation
agentsentinel version
```

## Quick Start

1. Start tmux: `tmux new-session -s agents`
2. Run your AI CLI inside tmux: `kiro` or `claude`
3. In another pane or terminal: `agentsentinel watch -v`

## Known Limitations

- macOS only (relies on tmux, which works on Linux but notifications are macOS-specific)
- Requires tmux - does not work with plain terminal windows

## Links

- Repository: https://github.com/plexusone/agentsentinel
- Issues: https://github.com/plexusone/agentsentinel/issues
