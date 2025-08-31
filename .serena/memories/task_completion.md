# Task Completion Workflow

## After Making Changes
1. Run `go fmt ./...` to format code
2. Run `golangci-lint run` to check for linting issues
3. Run `go test -v -race ./...` to ensure tests pass
4. Run `go mod tidy` to clean up dependencies
5. Run `go build` to verify the application compiles

## Before Committing
- Ensure all linting issues are resolved
- All tests must pass
- Code should be properly formatted
- Dependencies should be tidied

## CI/CD Notes
- GitHub Actions will automatically run build, test, and lint on push/PR
- Coverage reports are sent to Codecov
- Dependabot PRs are automatically approved and merged if builds pass