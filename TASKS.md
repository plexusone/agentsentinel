# AgentSentinel Tasks

Open items and recommendations from code review (2026-03-09).

## Testing

- [x] Add tests for `internal/notify` package (50.0% coverage)
- [x] Add tests for `internal/stats` package (88.7% coverage)
- [x] Add tests for `internal/config` package (73.9% coverage)
- [x] Add tests for `internal/watcher` package (85.7% coverage)
- [ ] Add integration tests for `cmd/` package

## Architecture

- [x] Add interface for `tmux.Client` to enable unit testing without real tmux (via `internal/watcher.TmuxClient`)
- [x] Add `context.Context` propagation through the watch loop for better cancellation handling
- [x] Extract watcher logic from `cmd/watch.go` into `internal/watcher` package (85.7% coverage)

## Features

- [ ] Add approval rate limiting to prevent rapid-fire approvals to the same pane
- [ ] Add health check endpoint for integration with monitoring systems
- [ ] Add metrics export (Prometheus format)

## Logging

Centralized context-based logging with ANSI colors.

### Completed

- [x] Add `github.com/lmittmann/tint` v1.1.3 for ANSI-colored slog output
- [x] Add `github.com/grokify/mogo` v0.73.4 for `slogutil.ContextWithLogger` / `LoggerFromContext`
- [x] Replace `slog.NewTextHandler` with `tint.NewHandler` in `cmd/watch.go`
- [x] Create context with logger via `slogutil.ContextWithLogger`
- [x] Pass context to `watcher.scan(ctx)` method
- [x] Remove `logger` field from `watcher` struct
- [x] Update `internal/stats.RecordApproval()` to accept `ctx` and use `LoggerFromContext`
- [x] Time format with seconds (`15:04:05`)

### Optional

- [ ] `internal/notify`: Add `ctx` parameter for future logging needs
- [ ] `cmd/status.go`: Convert `fmt.Print*` to structured slog calls
- [ ] `cmd/config.go`: Convert `fmt.Print*` to structured slog calls

## Code Quality

- [ ] Review error handling in `cmd/watch.go` - some errors are logged but continue silently
- [x] Add godoc comments to exported types and functions (all internal packages)
- [ ] Consider adding `//go:generate` for any code generation needs

## Documentation

- [x] Add architecture diagram to README (D2 → SVG in docs/)
- [ ] Add contributing guidelines (CONTRIBUTING.md)
- [x] Document pattern syntax for custom detection patterns (README.md)
- [x] Add CLAUDE.md for AI agents
- [x] Update CLAUDE.md with watcher package architecture and interfaces
