# Commands

AgentSentinel provides several commands for monitoring, testing, and configuration.

## watch

Monitor tmux panes and auto-approve tool requests.

```bash
agentsentinel watch [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--interval` | `-i` | `1s` | Scan interval |
| `--session` | `-s` | `""` | Watch specific tmux session only |
| `--dry-run` | | `false` | Detect but don't approve |
| `--block-danger` | | `true` | Block dangerous commands |
| `--lines` | | `30` | Lines to capture per pane |
| `--stats` | | `false` | Enable statistics tracking |
| `--notify` | | `false` | Enable macOS notifications |
| `--level` | `-l` | `info` | Log level (debug, info, warn, error) |
| `--verbose` | `-v` | `false` | Verbose output (same as --level debug) |

### Examples

```bash
# Watch all tmux panes
agentsentinel watch

# Watch with verbose logging
agentsentinel watch -v

# Watch a specific session
agentsentinel watch --session my-coding-session

# Faster scanning (500ms interval)
agentsentinel watch --interval 500ms

# Dry run mode (detect but don't approve)
agentsentinel watch --dry-run -v

# Enable stats and notifications
agentsentinel watch --stats --notify

# Set log level
agentsentinel watch --level debug
```

## status

Show tmux environment and available panes.

```bash
agentsentinel status
```

### Example Output

```
AgentSentinel Status
====================

tmux: installed
tmux server: running
current session: dev

Panes (4 total):

  %1 (session: $0, window: @0, index: 0)
  %2 (session: $0, window: @0, index: 1)
  %3 (session: $0, window: @1, index: 0)
  %4 (session: $1, window: @2, index: 0)

Run 'agentsentinel watch' to start monitoring for tool prompts.
```

## test

Verify prompt detection is working with sample prompts.

```bash
agentsentinel test
```

This runs detection against sample prompts and shows which patterns match.

## config

Manage the configuration file.

```bash
agentsentinel config <subcommand>
```

### Subcommands

| Subcommand | Description |
|------------|-------------|
| `path` | Show config file path |
| `show` | Show current configuration |
| `init` | Create example config file |
| `example` | Print example configuration |

### Examples

```bash
# Show config file path
agentsentinel config path
# ~/.agentsentinel.yaml

# Show current configuration
agentsentinel config show

# Create example config file
agentsentinel config init

# Print example configuration to stdout
agentsentinel config example
```

## version

Display version information.

```bash
agentsentinel version
```

## completion

Generate shell completion scripts.

```bash
agentsentinel completion <shell>
```

### Supported Shells

- `bash`
- `zsh`
- `fish`
- `powershell`

### Examples

```bash
# Bash
agentsentinel completion bash > /usr/local/etc/bash_completion.d/agentsentinel

# Zsh
agentsentinel completion zsh > "${fpath[1]}/_agentsentinel"

# Fish
agentsentinel completion fish > ~/.config/fish/completions/agentsentinel.fish
```

## Global Flags

These flags work with all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help for command |
| `--verbose` | `-v` | Enable verbose output |
