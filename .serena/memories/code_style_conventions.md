# Code Style and Conventions

## Go Code Style
- Standard Go formatting with `go fmt`
- Uses `gofumpt` for stricter formatting (enabled in .golangci.yml)

## Linting Rules (from .golangci.yml)
- `thelper` - enforces test helper function conventions
- `gofumpt` - stricter formatting than go fmt
- `tparallel` - detects inappropriate usage of t.Parallel()
- `unconvert` - removes unnecessary type conversions
- `unparam` - reports unused function parameters
- `wastedassign` - finds wasted assignment statements

## Bubble Tea Patterns (from main.go)
- Model struct contains application state
- Key bindings use charmbracelet/bubbles/key package
- Spinner styling with Lip Gloss color "205"
- Standard Bubble Tea methods: Init(), Update(), View()
- Error handling with custom errMsg type
- Quit on 'q', 'esc', or 'ctrl+c' keys