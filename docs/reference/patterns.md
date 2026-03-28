# Pattern Reference

AgentSentinel uses regex patterns to detect approval prompts and dangerous commands.

## Built-in Detection Patterns

AgentSentinel includes 19 built-in regex patterns for detecting approval prompts:

### Generic Y/N Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)\(y/n\)\s*$` | `Allow? (Y/n)` |
| `(?i)\[y/n\]\s*$` | `Continue? [y/n]` |
| `(?i)\(yes/no\)\s*$` | `Confirm? (yes/no)` |
| `(?i)\[yes/no\]\s*$` | `Proceed? [yes/no]` |

### Tool Request Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)allow\s*\?\s*\(y/n\)` | `Allow? (Y/n)` |
| `(?i)allow\s+tool` | `Allow tool execution` |
| `(?i)tool\s+request` | `Tool request pending` |
| `(?i)approve\s+tool` | `Approve tool use` |
| `(?i)proceed\s*\?` | `Proceed?` |
| `(?i)continue\s*\?\s*\(y/n\)` | `Continue? (Y/n)` |
| `(?i)execute\s*\?\s*\(y/n\)` | `Execute? (Y/n)` |
| `(?i)run\s+command\s*\?` | `Run command?` |

### Claude Code Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)allow\s+once` | `Allow once` |
| `(?i)allow\s+always` | `Allow always` |

### Codex CLI Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)sandbox\s+execution` | `Sandbox execution` |

### AWS Kiro Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)tool\s+use\s+\w+\s+requires\s+approval` | `tool use read requires approval` |
| `(?i)press\s+'y'\s+to\s+approve` | `press 'y' to approve` |

### Generic Patterns

| Pattern | Matches |
|---------|---------|
| `(?i)permission\s+required` | `Permission required` |
| `(?i)confirm\s+action` | `Confirm action` |

## Built-in Danger Patterns

These patterns block auto-approval when detected (13 patterns):

### Destructive Commands

| Pattern | Blocks |
|---------|--------|
| `(?i)rm\s+-rf` | `rm -rf /path` |
| `(?i)rm\s+-r\s+/` | `rm -r /` |
| `(?i)sudo\s+rm` | `sudo rm -rf` |
| `(?i)mkfs` | `mkfs.ext4 /dev/sda` |
| `(?i)dd\s+if=` | `dd if=/dev/zero of=/dev/sda` |

### Fork Bomb

| Pattern | Blocks |
|---------|--------|
| `(?i):\(\)\s*\{\s*:\|:\s*&\s*\}` | `:(){ :|:& };:` |

### Device Write

| Pattern | Blocks |
|---------|--------|
| `(?i)>\s*/dev/sd` | `> /dev/sda` |

### Permission Changes

| Pattern | Blocks |
|---------|--------|
| `(?i)chmod\s+-R\s+777\s+/` | `chmod -R 777 /` |
| `(?i)chown\s+-R.*\s+/` | `chown -R root /` |

### Remote Code Execution

| Pattern | Blocks |
|---------|--------|
| `(?i)curl.*\|\s*sh` | `curl url \| sh` |
| `(?i)curl.*\|\s*bash` | `curl url \| bash` |
| `(?i)wget.*\|\s*sh` | `wget url \| sh` |
| `(?i)wget.*\|\s*bash` | `wget url \| bash` |

## Custom Patterns

Add custom patterns in `~/.agentsentinel.yaml`:

```yaml
# Custom approval patterns
patterns:
  - "(?i)deploy\\s+to\\s+production\\?"
  - "(?i)apply\\s+changes\\?"
  - "my-tool-prompt"

# Custom danger patterns
danger_patterns:
  - "(?i)drop\\s+database"
  - "(?i)truncate\\s+table"
  - "(?i)delete\\s+from.*where\\s+1=1"
```

## Pattern Syntax

Patterns use Go's [regexp](https://pkg.go.dev/regexp/syntax) package (RE2 syntax).

### Common Syntax

| Syntax | Meaning | Example |
|--------|---------|---------|
| `(?i)` | Case-insensitive | `(?i)allow` matches `Allow`, `ALLOW`, `allow` |
| `\s+` | One or more whitespace | `tool\s+use` matches `tool use`, `tool  use` |
| `\s*` | Zero or more whitespace | `allow\s*\?` matches `allow?`, `allow ?` |
| `$` | End of line | `\(y/n\)$` only matches at line end |
| `\w+` | One or more word chars | `tool\s+\w+` matches `tool use`, `tool read` |
| `.*` | Any characters | `curl.*bash` matches `curl http://x \| bash` |
| `\|` | Literal pipe | `\|\s*sh` matches `\| sh` |

### Escaping Special Characters

| Character | Escaped | Example |
|-----------|---------|---------|
| `?` | `\?` | `proceed\?` |
| `(` `)` | `\(` `\)` | `\(y/n\)` |
| `.` | `\.` | `file\.txt` |
| `|` | `\|` | `curl\|bash` |
| `[` `]` | `\[` `\]` | `\[yes/no\]` |

### YAML Escaping

!!! warning "Double backslashes in YAML"

    In YAML strings, backslashes must be doubled:

    ```yaml
    patterns:
      - "(?i)deploy\\s+to\\s+production\\?"
    ```

## Testing Patterns

### Using the Test Command

```bash
agentsentinel test
```

This runs detection against sample prompts and shows which patterns match.

### Using Dry-Run Mode

```bash
agentsentinel watch --dry-run -v
```

This detects prompts but doesn't send approvals, letting you see what would be approved.

### Manual Testing

Test a regex pattern with Go:

```go
package main

import (
    "fmt"
    "regexp"
)

func main() {
    pattern := regexp.MustCompile(`(?i)deploy\s+to\s+production\?`)
    text := "Deploy to production? (y/n)"
    fmt.Println(pattern.MatchString(text)) // true
}
```
