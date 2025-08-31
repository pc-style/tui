# Suggested Commands

## Development Commands
- `go run main.go` - Run the application
- `go build` - Build the application
- `go test -v -race ./...` - Run tests with race detection
- `go test -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...` - Run tests with coverage (as used in CI)

## Code Quality Commands
- `golangci-lint run` - Run linting (requires golangci-lint to be installed)
- `go mod tidy` - Clean up go.mod and go.sum files
- `go fmt ./...` - Format Go code

## Release Commands
- `goreleaser build --snapshot --clean` - Build release candidates locally
- `goreleaser release --snapshot --clean` - Create a release snapshot

## Dependencies
- Install golangci-lint for linting: `brew install golangci-lint` (on macOS)
- Install goreleaser: `brew install goreleaser/tap/goreleaser` (on macOS)