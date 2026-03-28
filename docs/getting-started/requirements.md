# Requirements

## System Requirements

| Requirement | Version | Notes |
|-------------|---------|-------|
| **macOS** | 10.15+ | Primary supported platform |
| **tmux** | 3.0+ | Terminal multiplexer |
| **Go** | 1.21+ | Only needed for building from source |

## Installing tmux

If tmux is not already installed:

```bash
brew install tmux
```

Verify the installation:

```bash
tmux -V
# tmux 3.4
```

## tmux Basics

If you're new to tmux, here are the essential concepts:

### Sessions, Windows, and Panes

- **Session** - A collection of windows (like a workspace)
- **Window** - A full-screen view within a session (like a tab)
- **Pane** - A subdivision of a window (split view)

### Essential Commands

| Action | Command |
|--------|---------|
| Start new session | `tmux new-session -s name` |
| List sessions | `tmux list-sessions` |
| Attach to session | `tmux attach -t name` |
| Detach from session | `Ctrl+b d` |
| Split horizontally | `Ctrl+b "` |
| Split vertically | `Ctrl+b %` |
| Switch panes | `Ctrl+b arrow-key` |

### Why tmux?

AgentSentinel works by:

1. Listing all tmux panes via `tmux list-panes`
2. Reading pane content via `tmux capture-pane`
3. Sending keystrokes via `tmux send-keys`

This means your AI CLI must run inside a tmux pane for AgentSentinel to monitor and interact with it.

## Next Steps

- [Installation](installation.md) - Install AgentSentinel
- [Quick Start](quick-start.md) - Get up and running
