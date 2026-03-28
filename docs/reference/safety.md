# Safety Features

AgentSentinel includes several safety features to prevent accidental approval of dangerous commands.

## Dangerous Command Blocking

By default, AgentSentinel blocks auto-approval for commands that could cause significant damage.

### What Gets Blocked

| Category | Commands |
|----------|----------|
| **File deletion** | `rm -rf`, `rm -r /`, `sudo rm` |
| **Disk operations** | `mkfs`, `dd if=` |
| **Fork bomb** | `:(){ :\|:& };:` |
| **Device write** | `> /dev/sda` |
| **Permission changes** | `chmod -R 777 /`, `chown -R ... /` |
| **Remote execution** | `curl \| sh`, `curl \| bash`, `wget \| sh`, `wget \| bash` |

### When Blocking Occurs

When a dangerous command is detected:

1. The prompt is **not approved**
2. A warning is logged
3. If notifications are enabled, you receive an alert
4. The approval is recorded as "blocked" in stats

```
WARN dangerous command detected, skipping auto-approval pane=%3
```

### Disabling Danger Blocking

!!! danger "Not Recommended"

    Disabling danger blocking removes an important safety net.
    Only disable if you understand the risks.

Via CLI flag:

```bash
agentsentinel watch --block-danger=false
```

Via config:

```yaml
watch:
  block_danger: false
```

## Dry Run Mode

Test AgentSentinel without actually sending approvals:

```bash
agentsentinel watch --dry-run -v
```

In dry-run mode:

- Prompts are detected and logged
- Dangerous commands are flagged
- **No keystrokes are sent to tmux**

This is perfect for:

- Testing custom patterns
- Verifying detection works correctly
- Auditing what would be approved

### Example Output

```
INFO prompt detected pane=%1 type=Allow line="Allow? (Y/n)" blocked=false
INFO dry run: would approve pane=%1 count=1
```

## Duplicate Prevention

AgentSentinel tracks recently approved panes to prevent sending multiple approvals to the same prompt.

### How It Works

- After approving a pane, it's marked as "recently approved"
- For the next 5 seconds, that pane won't receive another approval
- This prevents double-approvals when the prompt hasn't cleared yet

### Why This Matters

Without duplicate prevention:

1. AgentSentinel detects a prompt and sends `y`
2. The AI CLI takes 500ms to process the approval
3. AgentSentinel scans again, sees the same prompt, sends another `y`
4. The extra `y` might approve something unintended

## Custom Danger Patterns

Add your own patterns to block:

```yaml
danger_patterns:
  - "(?i)drop\\s+database"
  - "(?i)truncate\\s+table"
  - "(?i)delete\\s+from.*where\\s+1=1"
  - "(?i)format\\s+c:"
  - "(?i)shutdown\\s+-h"
```

See [Pattern Reference](patterns.md) for syntax details.

## Notifications

Get alerted when dangerous commands are blocked:

```bash
agentsentinel watch --notify
```

Or in config:

```yaml
notifications:
  enabled: true
  sound: true
```

You'll receive a macOS notification when:

- A prompt is approved
- A dangerous command is blocked

## Statistics

Track blocked commands with stats:

```bash
agentsentinel watch --stats
```

On shutdown, you'll see blocked count:

```
AgentSentinel Statistics
========================
Total approvals: 15
Blocked (dangerous): 2
```

Enable logging to file for audit purposes:

```yaml
stats:
  enabled: true
  log_file: ~/agentsentinel-approvals.log
```

Each approval is logged as JSON:

```json
{"timestamp":"2024-01-15T10:30:45Z","pane_id":"%1","type":"Allow","line":"Allow? (Y/n)","blocked":false}
{"timestamp":"2024-01-15T10:31:02Z","pane_id":"%2","type":"Allow","line":"rm -rf / (Y/n)","blocked":true}
```

## Best Practices

1. **Always test with `--dry-run` first** - Verify detection works before going live

2. **Keep `--block-danger` enabled** - The default is there for a reason

3. **Review custom patterns carefully** - Overly broad patterns might match unintended prompts

4. **Use stats logging for audit** - Track what's being approved for later review

5. **Enable notifications** - Stay aware of blocked commands

6. **Monitor specific sessions** - Use `--session` to limit scope when testing
