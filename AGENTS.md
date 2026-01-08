# AGENTS.md - Developer Guide for Rehearse

## Project Overview

Rehearse is a Go CLI tool that analyzes GitHub Actions workflows without running them. It evaluates conditions, shows which jobs and steps would execute, and helps debug workflows locally.

**Module**: `github.com/telton/rehearse`
**Go Version**: 1.25.3
**Main Command**: `rehearse dryrun <workflow-file>`

## Essential Commands

### Build and Run
```bash
# Run without building
go run . dryrun <workflow-file>

# Build binary
go build -o rehearse .

# Example usage
go run . dryrun testdata/ci.yaml
go run . dryrun .github/workflows/go.yaml --event=pull_request --ref=feature-branch
```

### Testing and Linting
```bash
# Run tests (currently no test files exist)
go test -race -v ./...

# Verify dependencies
go mod verify
go mod tidy -diff

# Lint with golangci-lint (uses project config)
golangci-lint run --timeout=5m

# Security scan
govulncheck ./...
```

## Project Structure

```
rehearse/
├── main.go                 # Entry point, calls cmds.Execute()
├── cmds/                   # CLI command definitions
│   ├── root.go            # Root command setup using urfave/cli/v3
│   └── dryrun.go          # Main dryrun command implementation
├── workflow/              # Core workflow analysis logic
│   ├── types.go           # Workflow data structures
│   ├── parser.go          # YAML parsing for workflow files
│   ├── context.go         # GitHub Actions context simulation
│   ├── evaluator.go       # Expression evaluation (if conditions)
│   ├── analyzer.go        # Main analysis logic
│   ├── render.go          # Output formatting with lipgloss
│   ├── git.go             # Git repository integration
│   └── tokenizer.go       # Expression tokenization
├── testdata/              # Test workflow files
└── .github/workflows/     # Project CI configuration
```

## Code Patterns and Conventions

### CLI Framework
- Uses **urfave/cli/v3** for command structure
- Commands defined in `cmds/` package with `*cli.Command` structs
- Root command in `cmds/root.go`, subcommands in separate files

### Error Handling
```go
// Standard pattern throughout codebase
if err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

### YAML Unmarshaling
- Uses `github.com/goccy/go-yaml` for YAML parsing
- Custom `UnmarshalYAML` methods for flexible field types (see `RunsOn`, `Needs` in types.go)
- Handles both string and array formats for GitHub Actions syntax

### Struct Patterns
```go
// Workflow structs use yaml tags
type Job struct {
    Name   string            `yaml:"name"`
    RunsOn RunsOn            `yaml:"runs-on"`
    If     string            `yaml:"if"`
    // ...
}
```

### Context Lookup Pattern
- Path-based lookup system: `"github.ref"`, `"env.VAR_NAME"`, `"secrets.TOKEN"`
- Implemented in `workflow/context.go` with recursive map traversal

## Key Dependencies

- **CLI**: `github.com/urfave/cli/v3` - Command-line interface
- **YAML**: `github.com/goccy/go-yaml` - YAML parsing
- **Styling**: `github.com/charmbracelet/lipgloss` - Terminal UI styling

## Linting Configuration

Uses `.golangci.yaml` with strict settings:
- **Enabled linters**: gosec, errcheck, errorlint, govet, ineffassign, revive, sloglint, staticcheck, testifylint, unconvert, unused, usetesting, whitespace
- **Formatters**: gofmt with simplification, goimports with local prefix `github.com/telton`
- **Special rules**: `interface{}` → `any` rewrite rule

## CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/go.yaml`):
1. **test**: `go test -race -v ./...`, `go mod verify`, `go mod tidy -diff`
2. **lint**: `golangci-lint run --timeout=5m` (v2.4.0)
3. **govulncheck**: Security vulnerability scanning

Triggers on push/PR to main for Go files, go.mod/sum, and golangci config.

## Development Guidelines

### Adding New Commands
1. Create new file in `cmds/` package
2. Define `*cli.Command` struct with proper flags/arguments
3. Add to `rootCmd.Commands` slice in `cmds/root.go`

### Workflow Analysis Extension
- Core logic in `workflow/analyzer.go`
- Context building in `workflow/context.go`
- Expression evaluation in `workflow/evaluator.go`
- Add new GitHub Actions features by extending types in `workflow/types.go`

### Testing
- No test files currently exist
- When adding tests, follow Go conventions
- Integration tests should use `testdata/` directory for workflow files

## Important Notes

- **No Tests**: Project currently has no test coverage
- **Git Integration**: Automatically detects git repository information for context
- **Expression Evaluation**: Supports GitHub Actions expression syntax (`${{ }}`)
- **Flexible YAML**: Handles both string and array formats for `runs-on`, `needs`, etc.
- **Local Prefix**: Use `github.com/telton` import prefix for internal packages
- **Error Messages**: Always include context in error wrapping

## Memory Commands

When working on this project, remember:
```bash
# Test command
go test -race -v ./...

# Lint command  
golangci-lint run --timeout=5m

# Build command
go build -o rehearse .

# Run example
go run . dryrun testdata/ci.yaml
```