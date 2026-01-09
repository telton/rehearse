# rehearse

[![Go Reference](https://pkg.go.dev/badge/github.com/telton/rehearse.svg)](https://pkg.go.dev/github.com/telton/rehearse)
[![CI](https://github.com/telton/rehearse/actions/workflows/go.yaml/badge.svg)](https://github.com/telton/rehearse/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/telton/rehearse)](https://goreportcard.com/report/github.com/telton/rehearse)

**Practice before the real thing** - A CLI tool for analyzing and running GitHub Actions workflows locally.

Rehearse helps you develop, debug, and test GitHub Actions workflows without committing changes or running them in CI. It provides local workflow execution with Docker and comprehensive dry-run analysis to understand what would happen before your workflows run.

## Features

- 🔍 **Dry-run analysis** - See which jobs and steps would execute without running them
- 🐳 **Local execution** - Run workflows locally using Docker containers
- 📊 **Condition evaluation** - Understand complex workflow conditions and job dependencies  
- 🎯 **Event simulation** - Test different GitHub events (push, pull_request, etc.)
- 🔐 **Secret injection** - Provide secrets for local testing
- 📝 **Multiple output formats** - JSON and text output for integration
- ⚡ **Fast feedback** - Debug workflows without CI round trips

## Installation

### From Source

```bash
go install github.com/telton/rehearse@latest
```

### Build Locally

```bash
git clone https://github.com/telton/rehearse.git
cd rehearse
go build -o bin/rehearse .
```

## Quick Start

```bash
# Analyze a workflow to see what would run
rehearse dryrun .github/workflows/ci.yaml

# List all workflows in your repository
rehearse list

# Run a workflow locally with Docker
rehearse run .github/workflows/ci.yaml

# Test with different events and secrets
rehearse dryrun .github/workflows/deploy.yaml \
  --event=release \
  --secret="DEPLOY_TOKEN=your-token"
```

## Commands

### `rehearse dryrun`

Analyze a workflow without executing it. Shows which jobs and steps would run based on conditions and context.

```bash
rehearse dryrun [options] workflow-file
```

**Options:**
- `--event, -e` - Event type to simulate (default: "push")
- `--ref, -r` - Git ref to use (defaults to current branch)  
- `--secret, -s` - Secrets in KEY=VALUE format (can be repeated)

**Examples:**
```bash
# Basic analysis
rehearse dryrun .github/workflows/ci.yaml

# Simulate pull request event
rehearse dryrun .github/workflows/pr.yaml --event=pull_request

# Test with specific branch and secrets
rehearse dryrun .github/workflows/deploy.yaml \
  --ref=refs/heads/production \
  --secret="API_KEY=test123" \
  --secret="DB_PASSWORD=secret"
```

### `rehearse list`

List all available workflows in your repository.

```bash
rehearse list [options]
```

**Options:**
- `--dir, -d` - Directory to search (default: ".github/workflows")
- `--format, -f` - Output format: "text" or "json" (default: "text")
- `--pretty-print` - Enable pretty-print formatting for JSON

**Examples:**
```bash
# List workflows in default directory
rehearse list

# Search custom directory with JSON output
rehearse list --dir=workflows --format=json --pretty-print

# List workflows in current directory
rehearse list --dir=.
```

### `rehearse run`

Execute a workflow locally using Docker containers. **Requires Docker to be installed and running.**

```bash
rehearse run [options] workflow-file
```

**Options:**
- `--event, -e` - Event type to simulate (default: "push")
- `--ref, -r` - Git ref to use (defaults to current branch)
- `--secret, -s` - Secrets in KEY=VALUE format (can be repeated)
- `--working-dir` - Working directory for execution (default: current directory)
- `--pull` - Always pull Docker images before running
- `--cleanup` - Clean up containers and volumes after execution

**Examples:**
```bash
# Run workflow locally
rehearse run .github/workflows/test.yaml

# Run with cleanup and always pull images
rehearse run .github/workflows/ci.yaml --pull --cleanup

# Run with custom working directory and secrets
rehearse run ./workflows/deploy.yaml \
  --working-dir=/tmp/workspace \
  --secret="DEPLOY_KEY=xyz" \
  --cleanup
```

## Global Options

All commands support these global options:

- `--log-level, -l` - Set log level: "debug", "info", "warn", "error" (default: "info")
- `--help, -h` - Show help

You can also set the log level using the `REHEARSE_LOG_LEVEL` environment variable.

## Supported Features

Rehearse supports most GitHub Actions workflow features:

### Workflow Syntax
- ✅ Jobs with dependencies (`needs`)
- ✅ Conditional execution (`if` statements)  
- ✅ Environment variables (`env`)
- ✅ Multiple runner types (`runs-on`)
- ✅ Workflow triggers and events
- ✅ Job and step-level configuration

### Steps
- ✅ Shell commands (`run`)
- ✅ GitHub Actions (`uses`)
  - ✅ Local actions (`./path/to/action`)
  - ✅ Repository actions (`actions/checkout@v4`)
  - ✅ Docker actions (`docker://alpine:latest`)
- ✅ Step conditions and environment variables
- ✅ Step outputs and job outputs

### Actions
- ✅ Docker-based actions
- ✅ Node.js actions (16, 20)
- ✅ Composite actions
- ✅ Action inputs and outputs
- ✅ Local action development

### Context & Expressions
- ✅ GitHub context (`github.*`)
- ✅ Environment variables (`env.*`)
- ✅ Job outputs (`needs.*`)
- ✅ Step outputs (`steps.*`)
- ✅ Expression evaluation (`${{ }}`)

## Examples

### Basic Workflow Analysis

```bash
# Check what would run on push to main
rehearse dryrun .github/workflows/ci.yaml

# Output shows job execution plan:
# ╭───────────────────────────────╮
# │ [OK] Job: test                │
# │ runs-on: ubuntu-latest        │
# │                               │
# │   [OK] Checkout code          │
# │   [OK] Setup Go               │
# │   [OK] Run tests              │
# ╰───────────────────────────────╯
```

### Testing Different Events

```bash
# Test pull request workflow
rehearse dryrun .github/workflows/pr-check.yaml --event=pull_request

# Test release workflow  
rehearse dryrun .github/workflows/release.yaml --event=release
```

### Local Development Workflow

```bash
# 1. List available workflows
rehearse list

# 2. Analyze before running
rehearse dryrun .github/workflows/test.yaml --secret="API_KEY=test123"

# 3. Run locally for testing
rehearse run .github/workflows/test.yaml --secret="API_KEY=test123" --cleanup

# 4. Debug with verbose logging
rehearse --log-level=debug run .github/workflows/debug.yaml
```

### JSON Output for Integration

```bash
# Get workflow list as JSON
rehearse list --format=json --pretty-print

# Pipe to jq for processing
rehearse list --format=json | jq '.workflows[] | select(.name | contains("test"))'
```

## Requirements

- **Go 1.25+** for building from source
- **Docker** (for `run` command) - Must be installed and running
- **Git** - Used for repository context and action cloning

## Configuration

### Environment Variables

- `REHEARSE_LOG_LEVEL` - Set default log level (debug, info, warn, error)

### Git Integration

Rehearse automatically detects your git repository context:
- Current branch and commit SHA
- Repository owner and name (from remote origin)
- Git actor information

This context is used to simulate the GitHub environment that your workflows would see.

## Development

### Project Structure

```
rehearse/
├── main.go              # CLI entry point
├── cmds/               # Command definitions
│   ├── root.go         # Root command and global flags
│   ├── dryrun.go       # Dry-run analysis command
│   ├── list.go         # Workflow listing command
│   └── run.go          # Local execution command
├── workflow/           # Core workflow engine
│   ├── parser.go       # YAML parsing
│   ├── analyzer.go     # Workflow analysis
│   ├── executor.go     # Local execution
│   ├── context.go      # GitHub context simulation
│   └── evaluator.go    # Expression evaluation
└── testdata/           # Integration test workflows
```

### Building and Testing

```bash
# Build
go build -o bin/rehearse .

# Run tests
go test -race -v ./...

# Run integration tests only  
go test -v -race ./integration_test.go

# Lint code
golangci-lint run --timeout=5m

# Run benchmarks
go test -bench=. ./integration_test.go -benchtime=3s
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting (`go test ./... && golangci-lint run`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/telton/rehearse/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/telton/rehearse/discussions)
- 📖 **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/telton/rehearse)

## Acknowledgments

- Inspired by [act](https://github.com/nektos/act) for local GitHub Actions execution
- Built with [urfave/cli](https://github.com/urfave/cli) for the command-line interface
- Uses [lipgloss](https://github.com/charmbracelet/lipgloss) for beautiful terminal output