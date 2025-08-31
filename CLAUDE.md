# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Bubble Tea application template for creating terminal user interface (TUI) applications in Go. It uses the Charm ecosystem of tools for building interactive CLI applications.

## Tech Stack

- **Go 1.23.0** - Primary language
- **Bubble Tea** - TUI framework for terminal applications
- **Bubbles** - Pre-built UI components
- **Lip Gloss** - Styling library for terminal apps

## Essential Commands

### Development
```bash
go run main.go              # Run the application
go build                    # Build the executable
go test -v -race ./...      # Run tests with race detection
```

### Code Quality (run after making changes)
```bash
go fmt ./...                # Format Go code
golangci-lint run           # Run linting (requires golangci-lint)
go mod tidy                 # Clean up dependencies
```

### Dependencies
- Install golangci-lint: `brew install golangci-lint` (macOS)
- Install goreleaser: `brew install goreleaser/tap/goreleaser` (macOS)

## Architecture

### Bubble Tea Pattern
The application follows standard Bubble Tea architecture:
- **Model struct**: Contains application state
- **Init()**: Returns initial commands
- **Update()**: Handles messages and updates state  
- **View()**: Renders the current state

### Key Components
- Spinner animation using `bubbles/spinner`
- Key bindings with `bubbles/key`
- Styling with `lipgloss`
- Standard quit keys: 'q', 'esc', 'ctrl+c'

## Linting Configuration

Uses golangci-lint with these enabled linters:
- `gofumpt` - Stricter formatting than go fmt
- `thelper` - Test helper function conventions
- `tparallel` - t.Parallel() usage detection
- `unconvert` - Removes unnecessary type conversions
- `unparam` - Reports unused parameters
- `wastedassign` - Finds wasted assignments

## Release Process

GoReleaser is configured for automated releases:
- Builds for first-class Go platforms
- Generates changelog from conventional commits
- Triggered via GitHub Actions on tags