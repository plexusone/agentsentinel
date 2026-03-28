# AgentSentinel

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

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

[go-ci-svg]: https://github.com/plexusone/agentsentinel/actions/workflows/go-ci.yaml/badge.svg?branch=main
[go-ci-url]: https://github.com/plexusone/agentsentinel/actions/workflows/go-ci.yaml
[go-lint-svg]: https://github.com/plexusone/agentsentinel/actions/workflows/go-lint.yaml/badge.svg?branch=main
[go-lint-url]: https://github.com/plexusone/agentsentinel/actions/workflows/go-lint.yaml
[goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/agentsentinel
[goreport-url]: https://goreportcard.com/report/github.com/plexusone/agentsentinel
[docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/agentsentinel
[docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/agentsentinel
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://github.com/plexusone/agentsentinel/blob/master/LICENSE
