# AGENTS.md - Developer Guide for Rehearse

## Project Overview

Rehearse is a Go CLI tool that analyzes GitHub Actions workflows without running them. It evaluates conditions, shows which jobs and steps would execute, and helps debug workflows locally.

**Module**: `github.com/telton/rehearse`
**Go Version**: 1.25.5
**Main Command**: `rehearse dryrun <workflow-file>`

## Logging Configuration

The project uses structured logging with `log/slog`:

```bash
# Set log level via flag (debug, info, warn, error)
rehearse --log-level=debug list

# Set log level via environment variable
REHEARSE_LOG_LEVEL=warn rehearse dryrun workflow.yaml
```

- **Default level**: info
- **Structured logging**: Uses key-value pairs for better parsing
- **UI vs Logs**: UI output uses fmt.Print*, operational logs use logger package

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
│   ├── root.go            # Root command setup using urfave/cli/v3, logger config
│   ├── dryrun.go          # Main dryrun command implementation
│   ├── list.go            # List workflows command
│   └── run.go             # Execute workflows command
├── internal/              # Internal packages
│   └── logger/            # Structured logging with slog
├── ui/                    # Terminal UI components and styling
├── workflow/              # Core workflow analysis logic
│   ├── types.go           # Workflow data structures
│   ├── parser.go          # YAML parsing for workflow files
│   ├── context.go         # GitHub Actions context simulation
│   ├── evaluator.go       # Expression evaluation (if conditions)
│   ├── analyzer.go        # Main analysis logic
│   ├── render.go          # Output formatting with lipgloss
│   ├── run_render.go      # Execution output rendering
│   ├── executor.go        # Workflow execution engine
│   ├── step_executors.go  # Step execution implementations
│   ├── git.go             # Git repository integration
│   └── tokenizer.go       # Expression tokenization
├── testdata/              # Test workflow files
└── .github/workflows/     # Project CI configuration
```

## Code Patterns and Conventions

### Code Style Guidelines

**DO NOT USE:**
- **Emojis**: Never use emojis in code, comments, or UI output (use text alternatives like `[OK]`, `[FAIL]`, `[WARN]`)
- **Needless Comments**: Avoid obvious comments that just restate what the code does

**Bad Examples:**
```go
// 🚫 DON'T: Emojis in UI output
fmt.Println("✅ Success!")

// 🚫 DON'T: Obvious comments
// Increment counter by 1
counter++

// 🚫 DON'T: Explaining what code does
// Loop through all files
for _, file := range files {
```

**Good Examples:**
```go
// ✅ DO: Text alternatives for status
fmt.Println("[OK] Success!")

// ✅ DO: Comments explaining why, not what
// Use buffered channel to prevent goroutine leaks during shutdown
resultCh := make(chan Result, 1)

// ✅ DO: Comments for complex business logic
// GitHub Actions allows both string and array formats for runs-on
func (r *RunsOn) UnmarshalYAML(value *yaml.Node) error {
```

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

- **CLI**: `github.com/urfave/cli/v3` - Command-line interface with flag support
- **YAML**: `github.com/goccy/go-yaml` - YAML parsing for workflow files
- **Styling**: `github.com/charmbracelet/lipgloss` - Terminal UI styling and colors
- **Logging**: `log/slog` (standard library) - Structured logging with levels

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
- Use test files with `_test.go` suffix following Go conventions
- Integration tests should use `testdata/` directory for workflow files
- Logger tests included in `internal/logger/logger_test.go`

### Logging vs UI Output
**Use structured logging (`logger` package) for:**
- Debug information and execution flow tracking
- Warning messages (container cleanup failures, parsing errors)
- Error conditions that don't affect user-facing output
- Operational events for troubleshooting

**Use `fmt.Print*` statements for:**
- User-facing command output (list results, analysis output)
- Formatted UI components (headers, status messages, tables)
- Final rendered results that users expect to see

```go
// ✅ DO: Use logger for operational messages
logger.Warn("Failed to stop container", "container_id", containerID, "error", err)
logger.Debug("Starting workflow analysis", "workflow", workflowName)

// ✅ DO: Use fmt.Print for UI output
fmt.Println(status.Render())
fmt.Printf("%s: %s\n", f.Filename, f.WorkflowName)
```

## Important Notes

- **No Emojis**: Project uses text alternatives (`[OK]`, `[FAIL]`, `[DOCKER]`) instead of emojis
- **Structured Logging**: Uses `log/slog` with configurable levels via `--log-level` flag or `REHEARSE_LOG_LEVEL` env var
- **Git Integration**: Automatically detects git repository information for context
- **Expression Evaluation**: Supports GitHub Actions expression syntax (`${{ }}`)
- **Flexible YAML**: Handles both string and array formats for `runs-on`, `needs`, etc.
- **Local Prefix**: Use `github.com/telton` import prefix for internal packages
- **Error Messages**: Always include context in error wrapping

## Memory Commands

When working on this project, remember:
```bash
# Test command - runs all tests including comprehensive integration tests
go test -race -v ./...

# Test command (short mode) - skips stress tests for faster feedback
go test -race -v ./... -short

# Integration tests only - comprehensive workflow testing with parallel execution
go test -v -race ./integration_test.go

# Lint command  
golangci-lint run --timeout=5m

# Build command
go build -o bin/rehearse .

# Run example
go run . dryrun testdata/ci.yaml

# Test with debug logging
bin/rehearse --log-level=debug list

# Use environment variable for logging
REHEARSE_LOG_LEVEL=warn bin/rehearse dryrun workflow.yaml

# Benchmark tests - performance testing for parsing and analysis
go test -bench=. ./integration_test.go -benchtime=3s
```

## Integration Tests

The project includes comprehensive integration tests (`integration_test.go`) that:

- **Auto-discover** all workflow files in `testdata/` recursively
- **Test parsing** of all workflow files with parallel execution (`t.Parallel()`)
- **Test analysis** of valid workflows to ensure they execute correctly
- **Categorize tests** by directory (basic, features, errors, root)
- **Expression testing** with different contexts (main branch, feature branch, pull request)
- **Stress testing** with 50 parallel iterations for thread safety
- **Benchmark tests** for performance monitoring
- **Error handling** for semantically invalid workflows (in `testdata/errors/`)

**Test categories in testdata:**
- `testdata/basic/` - Simple workflow examples
- `testdata/features/` - Advanced feature demonstrations
- `testdata/errors/` - Error scenarios and edge cases
- `testdata/*.yaml` - Root level workflows

All tests use `t.Parallel()` for maximum performance and to verify thread safety.