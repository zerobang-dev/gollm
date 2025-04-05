# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands
- Build: `go build ./...`
- Run: `go run main.go`
- Test (all): `go test ./...`
- Test (single): `go test ./path/to/package -run TestName`
- Lint: `golangci-lint run`
- Format: `gofmt -w .`

## Code Style Guidelines
- **Imports**: Group standard library imports first, then third-party, then local packages
- **Formatting**: Follow Go standard format (gofmt)
- **Types**: Use strong typing, avoid interface{} when possible
- **Naming**: Follow Go conventions (CamelCase for exported, camelCase for private)
- **Error Handling**: Always check errors, return them up the call stack
- **Comments**: Use godoc style comments for exported functions and types
- **Concurrency**: Use channels and goroutines for concurrency, avoid shared memory
- **LLM Integration**: Follow best practices for token usage, context window management