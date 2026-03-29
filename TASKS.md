# AgentSentinel Tasks

Tracking open items and completed work.

**Current Version:** v0.2.0 (2026-03-28)

## Testing

- [x] Add tests for `internal/notify` package (50.0% coverage)
- [x] Add tests for `internal/stats` package (88.7% coverage)
- [x] Add tests for `internal/config` package (73.9% coverage)
- [x] Add tests for `internal/watcher` package (85.3% coverage)
- [ ] Add integration tests for `cmd/` package
- [ ] Increase `internal/tmux` coverage (currently 53.4%)

## Architecture

- [x] Add interface for `tmux.Client` to enable unit testing without real tmux (via `internal/watcher.TmuxClient`)
- [x] Add `context.Context` propagation through the watch loop for better cancellation handling
- [x] Extract watcher logic from `cmd/watch.go` into `internal/watcher` package

## Features

- [ ] Add approval rate limiting to prevent rapid-fire approvals to the same pane
- [ ] Add health check endpoint for integration with monitoring systems
- [ ] Add metrics export (Prometheus format)

## Logging

- [x] Add `github.com/lmittmann/tint` for ANSI-colored slog output
- [x] Add `github.com/grokify/mogo` for context-based logging
- [x] Add `--level` flag for log level control (debug, info, warn, error)
- [x] Context-based logging via `slogutil.ContextWithLogger`
- [ ] `internal/notify`: Add `ctx` parameter for future logging needs
- [ ] `cmd/status.go`: Convert `fmt.Print*` to structured slog calls
- [ ] `cmd/config.go`: Convert `fmt.Print*` to structured slog calls

## Code Quality

- [ ] Review error handling in `cmd/watch.go` - some errors are logged but continue silently
- [x] Add godoc comments to exported types and functions (all internal packages)
- [ ] Consider adding `//go:generate` for any code generation needs

## Documentation

- [x] Add architecture diagram (D2 → SVG in docs/)
- [x] Document pattern syntax for custom detection patterns (README.md)
- [x] Add CLAUDE.md for AI agents
- [x] Add MkDocs documentation site with Material theme
- [x] Add release notes in docs/releases/
- [ ] Add contributing guidelines (CONTRIBUTING.md)
- [ ] Deploy MkDocs to GitHub Pages
- [ ] Add documentation site badge to README

## Release

- [x] Prepare v0.2.0 changelog and release notes
- [ ] Tag and release v0.2.0
- [ ] Verify GoReleaser builds binaries correctly
