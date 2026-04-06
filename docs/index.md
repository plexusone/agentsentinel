# AgentSentinel

**Auto-approve tool requests for AI coding CLIs running in tmux.**

AgentSentinel monitors tmux panes for tool approval prompts from AI coding assistants and automatically responds with `y` to approve them. Designed for macOS with iTerm2 and tmux.

## Supported AI CLIs

| CLI | Prompt Format |
|-----|---------------|
| **AWS Kiro CLI** | `tool use read requires approval, press 'y' to approve` |
| **Claude Code** | `Allow? (Y/n)` |
| **Codex CLI** | `Proceed? (Y/n)` |
| **Gemini CLI** | `(Y/n)` prompts |

## Key Features

- **Automatic Approval** - Detects and approves tool requests without manual intervention
- **Multi-Agent Support** - Handles multiple concurrent AI agents across different tmux panes
- **Kiro Multi-Subagent TUI** - Special handling for Kiro's navigable approval list
- **Safety First** - Blocks dangerous commands like `rm -rf`, `sudo`, `curl|bash`
- **Configurable** - Custom patterns, notifications, and statistics tracking
- **Dry Run Mode** - Test detection without sending approvals

## How It Works

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

## Quick Links

- [Installation](getting-started/installation.md) - Get AgentSentinel running
- [Quick Start](getting-started/quick-start.md) - Step-by-step guide
- [Commands](usage/commands.md) - CLI reference
- [Configuration](usage/configuration.md) - YAML config options
- [Pattern Reference](reference/patterns.md) - Detection and danger patterns

## License

MIT
