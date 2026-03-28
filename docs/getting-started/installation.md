# Installation

## Option 1: Install from Go (Recommended)

Install directly to your `$GOPATH/bin`:

```bash
go install github.com/plexusone/agentsentinel@latest
```

Make sure `~/go/bin` is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

Verify the installation:

```bash
agentsentinel version
```

## Option 2: Build from Source

Clone and build locally:

```bash
git clone https://github.com/plexusone/agentsentinel.git
cd agentsentinel
go build -o agentsentinel .
```

Run from the current directory:

```bash
./agentsentinel version
```

Or move to a directory in your PATH:

```bash
mv agentsentinel /usr/local/bin/
```

## Shell Completion

Enable tab completion for your shell:

=== "Bash"

    ```bash
    agentsentinel completion bash > /usr/local/etc/bash_completion.d/agentsentinel
    ```

=== "Zsh"

    ```bash
    agentsentinel completion zsh > "${fpath[1]}/_agentsentinel"
    ```

=== "Fish"

    ```bash
    agentsentinel completion fish > ~/.config/fish/completions/agentsentinel.fish
    ```

## Verify Installation

Check that everything is working:

```bash
# Check version
agentsentinel version

# Check tmux connectivity
agentsentinel status
```

If `agentsentinel status` shows tmux information, you're ready to go.

## Next Steps

- [Quick Start](quick-start.md) - Start monitoring AI CLIs
