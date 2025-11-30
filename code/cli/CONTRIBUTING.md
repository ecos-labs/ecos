# Contributing to ecos CLI

This guide covers CLI-specific contribution guidelines. For general contribution information, see the [main CONTRIBUTING.md](../../CONTRIBUTING.md).

## Set up your machine

`ecos` CLI is written in [Go](https://go.dev/).

**Prerequisites:**

- [Go 1.24+](https://go.dev/doc/install)
- AWS CLI (for testing AWS plugins)
- Python 3.8+ and dbt (for testing transform plugins)

## Building

Clone the repository and navigate to the CLI directory:

```bash
git clone git@github.com:ecos-labs/ecos.git
cd ecos/code/cli
```

Install dependencies:

```bash
go mod download
```

Build the binary:

```bash
go build -o ecos main.go
./ecos --version
```

## Testing your changes

Before you commit, we suggest you run:

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

**Run specific tests:**

```bash
# Test specific package
go test ./cmd/
go test ./plugins/core/init/

# Test specific function
go test -run TestMyPlugin ./plugins/core/init/
```

## Pre-commit Hooks

We use [pre-commit](https://pre-commit.com/) to automatically check Go code quality before commits. For Go files, the hooks will:

- **Format code**: Run `go fmt` to ensure consistent formatting
- **Static analysis**: Run `go vet` to catch common mistakes
- **Linting**: Run `golangci-lint` for comprehensive code quality checks

**Installation:**
```bash
# From repository root
pip install pre-commit
pre-commit install
```

**Manual execution:**
```bash
# Run hooks on all files
pre-commit run --all-files

# Run hooks on staged files only (automatic on commit)
pre-commit run
```

**Note:** Pre-commit hooks run automatically when you commit Go files. If a hook fails, fix the issues and commit again.

**CI Enforcement:** The same pre-commit hooks are also run in CI on all pull requests. This ensures code quality standards are enforced even if local hooks are bypassed (e.g., using `git commit --no-verify`). Your PR will fail CI checks if the hooks don't pass.

For more information about all pre-commit hooks (including general file checks), see the [main CONTRIBUTING.md](../../CONTRIBUTING.md).

## Architecture Overview

The CLI uses a **command + plugin architecture**:

- **Commands** (`cmd/`) - Handle CLI interaction using Cobra
- **Plugins** (`plugins/core/`) - Contain provider/tool-specific logic
- **Registry** (`plugins/registry/`) - Self-registration system for plugins

**Key Pattern:** Plugins self-register using `init()` functions. Commands load plugins from the registry.

## Code Standards

**Go Style:**
- Follow standard Go conventions
- Use `gofmt` and `go vet`
- 4-space indentation
- Use named factory functions for plugins (not inline anonymous functions)

**Naming Conventions:**
```go
// Plugin types: <Provider><Source><Type>Plugin
type AWSCURInitPlugin struct{}
type AzureCostDestroyPlugin struct{}

// Factory functions: New<PluginName>
func NewAWSCUR(force bool, outputPath string) (types.InitPlugin, error)
func NewAwsCurDestroy() types.DestroyPlugin

// Registry names: lowercase with underscores
registry.RegisterInitPlugin("aws_cur", NewAWSCUR)
```

**Commits:** Use [Conventional Commits](https://www.conventionalcommits.org/) with `(cli)` scope:
- `feat(cli):` New feature
- `fix(cli):` Bug fix
- `docs(cli):` Documentation
- `refactor(cli):` Code refactoring
- `test(cli):` Adding or updating tests
- `chore(cli):` Maintenance tasks

**Example:**
```
feat(cli): add Azure Cost Management init plugin

This adds support for initializing ecos projects with Azure Cost Management
as the data source, including resource provisioning and dbt configuration.
```

## Adding a New Command

1. Create `cmd/<command>.go`:
```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringP("flag", "f", "", "Flag description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Add tests in `cmd/<command>_test.go`

## Adding a New Init Plugin

1. Create `plugins/core/init/<provider>.go`:
```go
type MyProviderInitPlugin struct {
    Config     *MyProviderInput
    Force      bool
    OutputPath string
}

// Implement all InitPlugin interface methods
func (p *MyProviderInitPlugin) Name() string { return "my-provider-init" }
func (p *MyProviderInitPlugin) ValidatePrerequisites() error { /* ... */ }
// ... implement other methods

// Factory function (required - init plugins need parameters)
func NewMyProvider(force bool, outputPath string) (types.InitPlugin, error) {
    return &MyProviderInitPlugin{
        Force:      force,
        OutputPath: outputPath,
        Config:     &MyProviderInput{},
    }, nil
}

// Self-register
func init() {
    registry.RegisterInitPlugin("my_provider", NewMyProvider)
}
```

2. Update `cmd/init.go` to include the new option in source selection

3. Add tests in `plugins/core/init/<provider>_test.go`

## Adding a New Destroy Plugin

1. Create `plugins/core/destroy/<provider>.go`:
```go
type MyProviderDestroyPlugin struct {
    // Plugin fields
}

// Implement DestroyPlugin interface
func (p *MyProviderDestroyPlugin) Name() string { return "my_provider" }
func (p *MyProviderDestroyPlugin) ValidatePrerequisites() error { /* ... */ }
func (p *MyProviderDestroyPlugin) LoadFromConfig(cfg *config.EcosConfig) error { /* ... */ }
func (p *MyProviderDestroyPlugin) DescribeDestruction() []types.DestroyResourcePreview { /* ... */ }
func (p *MyProviderDestroyPlugin) DestroyResources() ([]types.DestroyResourceResult, error) { /* ... */ }

// Factory function (no parameters - configured later via LoadFromConfig)
func NewMyProviderDestroy() types.DestroyPlugin {
    return &MyProviderDestroyPlugin{}
}

// Self-register
func init() {
    registry.RegisterDestroyPlugin("my_provider", NewMyProviderDestroy)
}
```

2. Add tests in `plugins/core/destroy/<provider>_test.go`

## Adding a New Transform Plugin

1. Create `plugins/core/transform/<tool>.go`:
```go
type MyToolTransformPlugin struct{}

// Implement TransformPlugin interface
func (p *MyToolTransformPlugin) Name() string { return "mytool" }
func (p *MyToolTransformPlugin) TransformEngine() string { return "mytool" }
func (p *MyToolTransformPlugin) ExecuteCommand(command string, args []string, config map[string]any) error { /* ... */ }
// ... implement other methods

// No registration needed - transform plugins are instantiated directly in cmd/transform.go
```

2. Update `cmd/transform.go` to instantiate the plugin:
```go
switch strings.ToLower(pluginName) {
case "dbt":
    plugin = &transform.DBTTransformPlugin{}
case "mytool":
    plugin = &transform.MyToolTransformPlugin{}
}
```

3. Add tests in `plugins/core/transform/<tool>_test.go`

## Submitting a pull request

For general PR guidelines, see the [main CONTRIBUTING.md](../../CONTRIBUTING.md).

**CLI-specific checklist before submitting:**

1. **Test your changes** - Run `go test ./...` and manual testing
2. **Format code** - Run `go fmt ./...`
3. **Update documentation** - Update `README.md` if adding new commands/features
4. **Fill out PR template** - Complete all sections

See [CLI README](README.md) for usage documentation.
