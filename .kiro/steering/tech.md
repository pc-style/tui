# Technology Stack

## Core Framework
- **Go 1.23.0+** - Primary programming language
- **Bubble Tea** - TUI framework for building interactive terminal applications
- **Bubbles** - Pre-built components for common TUI patterns
- **Lip Gloss** - Styling and layout library for terminal interfaces

## Build System & Tools
- **Go Modules** - Dependency management
- **GoReleaser** - Cross-platform binary releases and distribution
- **golangci-lint** - Code linting with custom configuration
- **gofumpt** - Enhanced Go code formatting

## Development Commands

### Building & Running
```bash
# Build application
go build -o bin/app .

# Run directly
go run main.go

# Build for release (uses GoReleaser)
goreleaser build --snapshot --clean
```

### Testing & Quality
```bash
# Run tests
go test ./...

# Run specific test
go test -run TestName ./path

# Lint code
golangci-lint run

# Format code
gofumpt -w .

# Update dependencies
go mod tidy
```

## Architecture Patterns
- **MVC Pattern**: Model struct, Update method (controller), View method (presentation)
- **Message-driven**: Uses Bubble Tea's Elm-inspired architecture
- **Component-based**: Leverages Bubbles components for reusable UI elements
- **Single binary**: Compiles to standalone executable with no external dependencies