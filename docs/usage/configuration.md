# Configuration

AgentSentinel can be configured via a YAML file at `~/.agentsentinel.yaml`.

## Configuration File

### Location

The default configuration path is:

```
~/.agentsentinel.yaml
```

Check the path with:

```bash
agentsentinel config path
```

### Creating a Config File

Generate an example configuration:

```bash
agentsentinel config init
```

Or print the example to stdout:

```bash
agentsentinel config example
```

## Configuration Options

### Full Example

```yaml
# Watch settings
watch:
  interval: 1s          # Scan interval (default: 1s)
  session: ""           # Filter to specific tmux session (default: all)
  lines: 30             # Lines to capture per pane (default: 30)
  block_danger: true    # Block dangerous commands (default: true)

# Custom detection patterns (Go regex syntax)
patterns:
  - "(?i)my-custom-prompt"
  - "(?i)deploy\\s+to\\s+production\\?"

# Custom danger patterns (Go regex syntax)
danger_patterns:
  - "(?i)drop\\s+database"
  - "(?i)truncate\\s+table"

# macOS notification settings
notifications:
  enabled: false        # Enable notifications (default: false)
  sound: true           # Play sound with notifications (default: true)
  title: "AgentSentinel"  # Notification title

# Statistics tracking
stats:
  enabled: false        # Enable stats tracking (default: false)
  log_file: ""          # JSON log file path (optional)
```

### Watch Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `interval` | duration | `1s` | How often to scan tmux panes |
| `session` | string | `""` | Limit to specific tmux session |
| `lines` | int | `30` | Lines to capture from each pane |
| `block_danger` | bool | `true` | Block dangerous commands |

### Custom Patterns

Add patterns to detect prompts not covered by built-in patterns:

```yaml
patterns:
  - "(?i)deploy\\s+to\\s+production\\?"
  - "(?i)apply\\s+changes\\?"
  - "my-custom-tool-prompt"
```

### Custom Danger Patterns

Add patterns to block from auto-approval:

```yaml
danger_patterns:
  - "(?i)drop\\s+database"
  - "(?i)truncate\\s+table"
  - "(?i)delete\\s+from.*where\\s+1=1"
```

### Notifications

Enable macOS notifications for approvals:

```yaml
notifications:
  enabled: true
  sound: true
  title: "AgentSentinel"
```

### Statistics

Track approval statistics:

```yaml
stats:
  enabled: true
  log_file: ~/agentsentinel-approvals.log
```

When enabled, stats are printed on shutdown (Ctrl+C). The log file records approvals in JSON format.

## CLI Flag Precedence

CLI flags override configuration file values:

```bash
# Config file has interval: 1s
# This overrides to 500ms
agentsentinel watch --interval 500ms
```

## Pattern Syntax

Patterns use Go's [regexp](https://pkg.go.dev/regexp/syntax) package (RE2 syntax):

| Syntax | Meaning |
|--------|---------|
| `(?i)` | Case-insensitive matching |
| `\s+` | One or more whitespace characters |
| `\s*` | Zero or more whitespace characters |
| `$` | End of line anchor |
| `\w+` | One or more word characters |
| `\?` | Literal question mark |
| `\(` `\)` | Literal parentheses |

### Example Patterns

```yaml
patterns:
  # Match "Deploy to production?" (case-insensitive)
  - "(?i)deploy\\s+to\\s+production\\?"

  # Match "Apply changes? [y/n]"
  - "(?i)apply\\s+changes\\?\\s*\\[y/n\\]"

  # Match exact string
  - "my-tool: confirm action"
```

!!! note "Escaping in YAML"

    In YAML, backslashes must be doubled: `\\s` instead of `\s`

## Testing Configuration

Test your patterns:

```bash
# Run the test command
agentsentinel test

# Or use dry-run with verbose output
agentsentinel watch --dry-run -v
```

## See Also

- [Pattern Reference](../reference/patterns.md) - Built-in patterns
- [Safety Features](../reference/safety.md) - Danger pattern details
