# Project Structure

## Root Directory Layout
```
├── .git/                   # Git version control
├── .kiro/                  # Kiro IDE configuration and steering
├── .vscode/                # VS Code workspace settings
├── bin/                    # Compiled binaries and executables
├── main.go                 # Single-file application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── .gitignore              # Git ignore patterns
├── .golangci.yml           # Linting configuration
├── .goreleaser.yaml        # Release automation config
├── AGENTS.md               # Development guidelines
├── LICENSE                 # Project license
└── README.md               # Project documentation
```

## Code Organization

### Single-File Architecture
- **main.go** - Contains the entire application in one file (~1400 lines)
- All types, functions, and logic are co-located for simplicity
- Follows Bubble Tea's recommended structure for small-to-medium applications

### Key Components in main.go
- **Types**: `model`, `Process`, `SystemStats`, `NetworkConnection`, etc.
- **Initialization**: `initialModel()`, `Init()` methods
- **Event Handling**: `Update()` method with message routing
- **Rendering**: `View()` method and render helper functions
- **System Integration**: OS-specific commands for process/network data

## Configuration Files

### Development Tools
- **.golangci.yml** - Enables specific linters (thelper, gofumpt, tparallel, etc.)
- **.goreleaser.yaml** - Cross-platform build targets and release automation
- **AGENTS.md** - Development commands and coding standards

### Build Outputs
- **bin/** directory contains compiled executables
- Multiple binary names suggest different build targets or versions

## Conventions
- Single-file approach keeps complexity low for template usage
- All styling uses lipgloss with consistent neon color scheme
- System calls are platform-aware (uses `ps`, `netstat`, `sysctl` on macOS)
- Error handling follows Go idioms with explicit error returns