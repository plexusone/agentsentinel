# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release
- `watch` command to monitor tmux panes for tool approval prompts
- `status` command to display tmux environment
- `test` command to verify prompt detection
- `version` command
- Support for AWS Kiro CLI, Claude Code, Codex CLI, Gemini CLI
- Dangerous command blocking (rm -rf, sudo, curl|bash, etc.)
- Dry-run mode for testing without approval
- Configurable scan interval
- Session filtering
- Verbose logging
