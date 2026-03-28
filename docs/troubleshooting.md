# Troubleshooting

Common issues and solutions when using AgentSentinel.

## tmux Issues

### "tmux is not installed"

Install tmux via Homebrew:

```bash
brew install tmux
```

### "tmux server is not running"

Start a tmux session first:

```bash
tmux new-session -s agents
```

Then run AgentSentinel.

### Can't see any panes

Check that tmux is running and has panes:

```bash
# List all tmux sessions
tmux list-sessions

# List all panes
tmux list-panes -a
```

If no panes are listed, start a tmux session first.

## Detection Issues

### Prompts not being detected

1. **Enable verbose logging**

    ```bash
    agentsentinel watch -v
    ```

2. **Check the prompt is in the visible area**

    AgentSentinel captures the last 30 lines by default. If the prompt scrolled up:

    ```bash
    agentsentinel watch --lines 50
    ```

3. **Run the test command**

    ```bash
    agentsentinel test
    ```

4. **Check the exact prompt format**

    The prompt may use a different format than expected. Run `agentsentinel status` and manually inspect the pane content:

    ```bash
    tmux capture-pane -t %1 -p | tail -30
    ```

5. **Add a custom pattern**

    If the prompt format isn't covered by built-in patterns:

    ```yaml
    # ~/.agentsentinel.yaml
    patterns:
      - "(?i)my-custom-prompt"
    ```

### Wrong prompts being detected

Use dry-run mode to see what's being detected:

```bash
agentsentinel watch --dry-run -v
```

If unwanted prompts are being matched, the built-in patterns may be too broad for your use case. Consider filing an issue.

## Approval Issues

### Approvals not working

1. **Verify the pane ID**

    ```bash
    agentsentinel status
    ```

2. **Test manually**

    Send a keystroke directly to verify tmux can interact with the pane:

    ```bash
    tmux send-keys -t %1 y Enter
    ```

3. **Check CLI focus**

    Some CLIs may require focus or specific terminal state. Ensure the CLI is waiting for input.

4. **Check for duplicate prevention**

    If you recently approved the same pane, AgentSentinel waits 5 seconds before re-approving. Wait and try again.

### Double approvals

This shouldn't happen with the default duplicate prevention. If it does:

1. Check if you're running multiple AgentSentinel instances
2. Increase the scan interval: `--interval 2s`

## Configuration Issues

### Config file not loading

1. **Check the path**

    ```bash
    agentsentinel config path
    ```

2. **Verify YAML syntax**

    ```bash
    agentsentinel config show
    ```

    If this errors, there's a syntax issue in your config file.

3. **Common YAML mistakes**

    - Missing spaces after colons: `interval:1s` → `interval: 1s`
    - Incorrect indentation (use spaces, not tabs)
    - Unescaped special characters in patterns

### Custom patterns not working

1. **Check YAML escaping**

    Backslashes must be doubled in YAML:

    ```yaml
    patterns:
      - "(?i)deploy\\s+to\\s+production\\?"  # Correct
      - "(?i)deploy\s+to\s+production\?"     # Wrong
    ```

2. **Test the pattern**

    ```bash
    agentsentinel test
    ```

3. **Use dry-run mode**

    ```bash
    agentsentinel watch --dry-run -v
    ```

## Performance Issues

### High CPU usage

Reduce scan frequency:

```bash
agentsentinel watch --interval 2s
```

### Slow response time

If approvals seem delayed:

1. Decrease scan interval: `--interval 500ms`
2. Reduce lines captured: `--lines 20`

## Notification Issues

### macOS notifications not appearing

1. **Check System Preferences**

    Go to System Preferences → Notifications and ensure "Script Editor" or terminal app has notification permissions.

2. **Test osascript directly**

    ```bash
    osascript -e 'display notification "Test" with title "AgentSentinel"'
    ```

3. **Check Do Not Disturb**

    Ensure Do Not Disturb is not enabled.

## Getting Help

If none of these solutions work:

1. **Check existing issues**

    [GitHub Issues](https://github.com/plexusone/agentsentinel/issues)

2. **File a new issue**

    Include:

    - AgentSentinel version: `agentsentinel version`
    - macOS version
    - tmux version: `tmux -V`
    - Verbose log output: `agentsentinel watch -v`
    - The exact prompt text you're trying to detect
