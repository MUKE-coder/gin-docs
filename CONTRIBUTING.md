# Contributing to Gin Docs

Thank you for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/MUKE-coder/gin-docs.git
cd gin-docs
go mod tidy
go test ./...
```

## Running the Demo

```bash
go run main.go
# Visit http://localhost:8080/docs
```

## Code Standards

- Run `go vet ./...` before submitting
- Run `go test ./...` and ensure all tests pass
- Add tests for new functionality
- Every exported function needs a doc comment
- No panics — return errors gracefully

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with a descriptive message
6. Push and open a PR

## Reporting Issues

Open an issue on GitHub with:
- Go version (`go version`)
- Gin version
- Steps to reproduce
- Expected vs actual behavior
