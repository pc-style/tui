# AGENTS.md - Development Guidelines

## Commands
- **Build:** `go build -o bin/app .` 
- **Run:** `go run main.go`
- **Test:** `go test ./...` (for single test: `go test -run TestName ./path`)
- **Lint:** `golangci-lint run` (uses .golangci.yml config)
- **Format:** `gofumpt -w .`
- **Dependencies:** `go mod tidy`

## Architecture
- Simple Bubble Tea TUI application with single main.go
- Uses Charm libraries: bubbletea (framework), bubbles (components), lipgloss (styling)
- MVC pattern: model struct, Update method (controller), View method (presentation)
- Built for cross-platform TUI applications with GoReleaser

## Code Style
- **Imports:** Standard library first, then third-party, blank line separated
- **Naming:** camelCase for unexported, PascalCase for exported
- **Error handling:** Return errors explicitly, use errMsg type for tea.Msg
- **Types:** Define custom types for clarity (e.g., `type errMsg error`)
- **Key bindings:** Use bubbles/key.NewBinding for consistent UX
- **Styling:** Use lipgloss for all visual styling, define colors consistently
