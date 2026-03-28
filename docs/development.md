# Development

Guide for developing and contributing to AgentSentinel.

## Prerequisites

- Go 1.21+
- tmux 3.0+
- golangci-lint (for linting)

## Getting Started

Clone the repository:

```bash
git clone https://github.com/plexusone/agentsentinel.git
cd agentsentinel
```

Build:

```bash
go build -o agentsentinel .
```

Run:

```bash
./agentsentinel version
```

## Development Commands

### Build

```bash
# Build binary
go build -o agentsentinel .

# Install to $GOPATH/bin
go install .
```

### Test

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Lint

```bash
# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

### Format

```bash
# Format code
gofmt -w .
```

## Project Structure

```
agentsentinel/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
├── internal/               # Internal packages
│   ├── watcher/            # Core watch logic
│   ├── config/             # Configuration
│   ├── detector/           # Pattern detection
│   ├── tmux/               # tmux integration
│   ├── stats/              # Statistics
│   └── notify/             # Notifications
├── docs/                   # Documentation
├── mkdocs.yml              # MkDocs config
├── go.mod                  # Go module
└── go.sum                  # Go dependencies
```

## Adding Detection Patterns

### Built-in Patterns

Add patterns in `internal/detector/detector.go`:

```go
// In NewDetector()
patterns: []*regexp.Regexp{
    regexp.MustCompile(`(?i)your-new-pattern`),
    // ...
}
```

### Adding Tests

Add test cases in `internal/detector/detector_test.go`:

```go
{
    name:     "Your new pattern",
    input:    "Your prompt text (Y/n)",
    wantType: PromptAllow,
    wantLine: "Your prompt text (Y/n)",
}
```

## Adding Danger Patterns

In `internal/detector/detector.go`:

```go
dangerPatterns: []*regexp.Regexp{
    regexp.MustCompile(`(?i)dangerous-command`),
}
```

## Adding a New Command

1. Create `cmd/newcmd.go`:

```go
package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
    Use:   "newcmd",
    Short: "Description of the command",
    RunE:  runNewCmd,
}

func init() {
    rootCmd.AddCommand(newCmd)
}

func runNewCmd(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Follow existing command structure (see `cmd/status.go` for a simple example).

## Modifying Watch Behavior

Core watch logic is in `internal/watcher/watcher.go`:

- `watcher.New()` - Create watcher with dependencies
- `watcher.Scan(ctx)` - Per-tick scan of all panes
- `wasRecentlyApproved()` - Duplicate prevention
- `markApproved()` - Track approved panes

The `cmd/watch.go` handles CLI setup and the main ticker loop.

## Testing Without tmux

The watcher package uses interfaces for dependency injection:

```go
type TmuxClient interface {
    ListPanes() ([]string, error)
    CapturePane(paneID string, lines int) (string, error)
    Approve(paneID string) error
    ApproveMultiple(paneID string, count int, delayMs int) error
}
```

Create mock implementations for testing:

```go
type mockTmuxClient struct {
    panes        []string
    paneContents map[string]string
}

func (m *mockTmuxClient) ListPanes() ([]string, error) {
    return m.panes, nil
}
// ... implement other methods
```

See `internal/watcher/watcher_test.go` for examples.

## Documentation

### Building Docs

Install MkDocs with Material theme:

```bash
pip install mkdocs-material
```

Serve locally:

```bash
mkdocs serve
```

Build static site:

```bash
mkdocs build
```

### Updating Architecture Diagram

Edit `docs/architecture.d2` and regenerate:

```bash
d2 docs/architecture.d2 docs/architecture.svg
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Keep functions focused and small
- Prefer returning errors over logging them

## Error Handling

Follow this priority:

1. **Return** - If the function can return an error
2. **Log** - Use context-based logging: `slogutil.LoggerFromContext(ctx)`
3. **Panic** - Only for programming errors / invariant violations

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(detector): add support for new CLI prompt
fix: resolve race condition in approval tracking
docs: update pattern reference
test: add integration tests for watcher
```

## Pull Request Guidelines

1. Create a feature branch from `main`
2. Write tests for new functionality
3. Ensure all tests pass: `go test ./...`
4. Ensure linting passes: `golangci-lint run`
5. Update documentation if needed
6. Submit PR with clear description

## Release Process

1. Update version in code
2. Update CHANGELOG
3. Create and push tag: `git tag v1.0.0 && git push --tags`
4. GitHub Actions will build and release
