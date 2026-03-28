# Quick Start

This guide gets you from zero to auto-approving AI tool requests in minutes.

## Step 1: Start tmux

Open iTerm2 and start a tmux session:

```bash
tmux new-session -s agents
```

You're now inside tmux.

## Step 2: Run Your AI CLI

Inside the tmux session, start your AI coding assistant:

```bash
# AWS Kiro
kiro

# Or Claude Code
claude

# Or Codex CLI
codex
```

## Step 3: Run AgentSentinel

You have three options for where to run AgentSentinel:

### Option A: In Another tmux Pane (Recommended)

Split the tmux window and run AgentSentinel in the new pane:

```bash
# Split horizontally
# Press: Ctrl+b "

# Or split vertically
# Press: Ctrl+b %

# In the new pane, run:
agentsentinel watch -v
```

### Option B: In a Separate iTerm2 Tab

Open a new iTerm2 tab (Cmd+T) and run AgentSentinel directly (not in tmux):

```bash
agentsentinel watch -v
```

### Option C: Background tmux Session

Create a detached session for AgentSentinel:

```bash
tmux new-session -d -s sentinel 'agentsentinel watch -v'
```

## Step 4: Watch It Work

Once running, AgentSentinel monitors all tmux panes and auto-approves tool requests:

```
┌─────────────────────────────────────┐
│ tmux pane 1: kiro                   │
│                                     │
│ > tool use read requires approval,  │
│   press 'y' to approve              │  ← Detected
│                                     │
├─────────────────────────────────────┤
│ tmux pane 2: agentsentinel watch -v │
│                                     │
│ INFO prompt detected pane=%1        │
│ INFO approved pane=%1               │  ← Approved
└─────────────────────────────────────┘
```

## Key Concept

!!! info "AgentSentinel monitors tmux from the outside"

    AgentSentinel uses tmux commands to read pane contents and send keystrokes.
    It doesn't run "inside" your AI CLI. It runs separately and watches all tmux panes.

## Quick Test

Before relying on AgentSentinel, test it:

```bash
# Check AgentSentinel can see your tmux panes
agentsentinel status

# Dry run to see what it would approve (without actually approving)
agentsentinel watch --dry-run -v
```

## Where to Run AgentSentinel

| Location | Pros | Cons |
|----------|------|------|
| **tmux pane** | See logs alongside agents | Uses a pane |
| **Separate iTerm2 tab** | Doesn't use tmux space | Switch tabs to see logs |
| **Background session** | Out of the way | Need to attach to see logs |

## Next Steps

- [Commands](../usage/commands.md) - Full CLI reference
- [Configuration](../usage/configuration.md) - Customize behavior
- [Multi-Agent Workflows](../usage/multi-agent.md) - Run multiple AI agents
