# Test Workflows

This directory contains example workflows for testing and demonstrating rehearse functionality.

## Basic Workflows

**`basic/hello.yaml`** - Simple single-job workflow with basic GitHub context usage
- Single job with multiple steps
- Demonstrates basic expressions (`${{ runner.os }}`, `${{ github.ref }}`)
- Good for initial testing

**`basic/multi-job.yaml`** - Multi-job workflow with dependencies and outputs  
- Job dependencies (`needs`)
- Job outputs and cross-job data passing
- Conditional deployment based on branch
- Tests job orchestration

**`ci.yaml`** - Complete CI pipeline example
- Multiple dependent jobs (lint, test, build, deploy, notify)
- Complex conditionals and job dependencies
- Demonstrates real-world CI/CD patterns

## Feature Demonstrations

**`features/conditionals.yaml`** - Job and step conditional execution
- Different trigger events (`push`, `pull_request`)
- Branch-based conditions
- Job-level and step-level `if` conditions
- Tests expression evaluation logic

**`features/environment-outputs.yaml`** - Environment variables and outputs
- Global, job-level, and step-level environment variables
- Dynamic environment variables (`GITHUB_ENV`)
- Step outputs (`GITHUB_OUTPUT`) 
- Cross-job output usage

**`features/expressions-demo.yaml`** - GitHub Actions expression evaluation
- Context variables (`github.*`)
- String functions (`contains`, `startsWith`, `endsWith`)
- Boolean logic and comparisons
- Step output references

**`features/actions.yaml`** - External action usage
- Common GitHub Actions (`checkout`, `setup-node`, `cache`)
- Action parameters and configuration  
- Mixed shell commands and actions

**`features/action-test.yaml`** - Action testing scenarios
- Various action usage patterns
- Action parameter validation

**`features/showcase.yaml`** - Comprehensive feature showcase
- Demonstrates multiple features in one workflow
- Complex real-world scenarios

## Error Cases

**`errors/invalid-workflow.yaml`** - Various error scenarios
- Syntax errors in expressions
- Missing job dependencies  
- Non-existent actions
- Failing steps
- Tests error handling and reporting

## Usage

Use these workflows to test different rehearse features:

```bash
# Basic functionality
rehearse dryrun testdata/basic/hello.yaml
rehearse run testdata/basic/hello.yaml

# Test conditionals with different events
rehearse dryrun testdata/features/conditionals.yaml --event=push
rehearse dryrun testdata/features/conditionals.yaml --event=pull_request

# Test expressions and outputs
rehearse run testdata/features/expressions-demo.yaml

# Test error handling
rehearse dryrun testdata/errors/invalid-workflow.yaml
```