# Multi-Agent Workflows

AgentSentinel handles multiple concurrent AI agents across different tmux panes, including special support for Kiro's multi-subagent TUI.

## Multiple Panes

Run different AI agents in separate tmux panes:

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

AgentSentinel scans all panes on each interval, detecting and approving prompts in any pane.

## Kiro Multi-Subagent TUI

AWS Kiro can run multiple subagents concurrently (typically 4). These appear in a navigable TUI list within a single pane:

```
┌─────────────────────────────────────────────┐
│ kiro                                        │
│                                             │
│ ↳ tool use read requires approval,         │
│   press 'y' to approve, 'n' to deny...     │
│ ↳ tool use write requires approval,        │
│   press 'y' to approve, 'n' to deny...     │
│ ↳ tool use execute requires approval,      │
│   press 'y' to approve, 'n' to deny...     │
│ ↳ tool use read requires approval,         │
│   press 'y' to approve, 'n' to deny...     │
└─────────────────────────────────────────────┘
```

### How AgentSentinel Handles This

When AgentSentinel detects Kiro's multi-subagent format:

1. **Counts pending prompts** - Identifies how many `tool use ... requires approval` lines exist
2. **Cycles through all items** - Sends `y` (approve) + `j` (navigate down) repeatedly
3. **Full coverage** - Performs count+1 iterations to ensure all items are approved regardless of cursor starting position

```bash
# Example log output
INFO kiro multi-subagent detected, cycling through all pane=%1 detected=4 cycles=4
INFO approved pane=%1 count=4
```

## Session Filtering

If you have multiple tmux sessions, you can limit monitoring to a specific one:

```bash
# Watch only the "agents" session
agentsentinel watch --session agents
```

Or in config:

```yaml
watch:
  session: "agents"
```

## Faster Scanning

For responsive approval with multiple agents, reduce the scan interval:

```bash
# Scan every 500ms
agentsentinel watch --interval 500ms
```

## Background with Logging

Run AgentSentinel in the background while working:

```bash
# Background with logging
agentsentinel watch -v > ~/agentsentinel.log 2>&1 &

# Check logs
tail -f ~/agentsentinel.log

# Stop it
pkill agentsentinel
```

Or use a dedicated tmux session:

```bash
# Create detached session
tmux new-session -d -s sentinel 'agentsentinel watch -v'

# Attach to check logs
tmux attach -t sentinel

# Detach again
# Press: Ctrl+b d
```

## Duplicate Prevention

AgentSentinel tracks recently approved panes to prevent sending multiple approvals to the same prompt. By default, it won't re-approve the same pane within 5 seconds.

This prevents issues when:

- The prompt hasn't cleared yet after approval
- Multiple scan cycles occur before the AI CLI processes the approval

## Statistics for Multiple Agents

Enable stats to track approvals across all agents:

```bash
agentsentinel watch --stats
```

On shutdown (Ctrl+C), you'll see:

```
AgentSentinel Statistics
========================
Total scans: 120
Total approvals: 15
Blocked (dangerous): 0

Approvals by pane:
  %1: 8
  %2: 5
  %3: 2

Approvals by type:
  Approve: 8
  Allow: 7
```
