# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands
- Build: `go build`
- Run tests: `go test -v`
- Run specific test: `go test -run TestMainWithSnapshots -v`
- Record new snapshots: `RECORD=true go test -v`
- Run the application: `./asana-tasks-sorter`
- Run with custom config: `./asana-tasks-sorter -config path/to/config.json`
- Run in dry-run mode: `./asana-tasks-sorter -dry-run`

## Code Style Guidelines
- **Formatting**: Use standard Go style (gofmt)
- **Imports**: Group standard library first, then project imports
- **Error Handling**: Check errors explicitly, propagate with descriptive context
- **Documentation**: Add comments for exported functions and types
- **Types**: Use explicit type definitions for domain objects
- **Naming**: CamelCase for exported identifiers, camelCase for unexported ones

## Project Structure
- Main package for CLI functionality
- Internal packages for API client and testing utilities
- Snapshot testing for API interactions without real calls
- Zero external dependencies - uses only Go standard library