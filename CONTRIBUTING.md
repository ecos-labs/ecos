# Contributing to ecos

Thank you for your interest in contributing to ecos! This guide will help you get started.

## Getting Started

### Prerequisites

- Python 3.8+ installed
- AWS CLI configured (for testing with Athena)
- Git installed

### Setting Up Your Development Environment

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/ecos-core.git
cd ecos-core

# Create a virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Install dbt packages
cd code/dbt && dbt deps && cd ../..

# Install pre-commit hooks (recommended)
pip install pre-commit
pre-commit install
```

### Pre-commit Hooks

We use [pre-commit](https://pre-commit.com/) to automatically check code quality before commits. The hooks will:

- **General checks**: Remove trailing whitespace, ensure files end with newlines, validate YAML/JSON, check for merge conflicts
- **Go code**: Format with `go fmt`, run `go vet`, and lint with `golangci-lint`
- **SQL code**: Lint and auto-fix with SQLFluff (dbt-aware)

**Installation:**
```bash
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

**Note:** Pre-commit hooks run automatically when you commit. If a hook fails, fix the issues and commit again.

**CI Enforcement:** The same pre-commit hooks are also run in CI on all pull requests. This ensures code quality standards are enforced even if local hooks are bypassed (e.g., using `git commit --no-verify`). Your PR will fail CI checks if the hooks don't pass.

### Making Your First Contribution

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following our code standards (see below)

3. **Test your changes**:
   ```bash
   # For dbt models
   cd code/dbt
   dbt run --select your_model+
   dbt test --select your_model+

   # For Python code
   cd code/mcp
   pytest
   ```

4. **Commit and push**:
   ```bash
   git commit -m "feat: add your feature"
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request** on GitHub

## Code Standards

### SQL Style Guide

We use [sqlfluff](https://docs.sqlfluff.com/) to enforce consistent SQL formatting. Our configuration uses the Athena dialect and follows these conventions:

**Key Rules:**
- Lowercase keywords, identifiers, and functions
- Leading commas (comma at the start of the line)
- 4-space indentation
- Maximum line length: 120 characters
- Explicit table and column aliases

**Example:**
```sql
select
    column_a
    , column_b
    , sum(column_c) as total_column_c
from {{ ref('source_table') }} as st
where column_a is not null
group by column_a, column_b
```

**Linting Commands:**
```bash
# Check for linting issues
cd code/dbt
sqlfluff lint models/your_model.sql

# Auto-fix linting issues
sqlfluff fix models/your_model.sql

# Lint all SQL files
sqlfluff lint .

# Check specific rules
sqlfluff lint models/your_model.sql --rules L001,L003
```

**Configuration:** See `.sqlfluff` in `code/dbt/` for full configuration details.

### Python Style Guide

Follow [PEP 8](https://pep8.org/) with these additions:
- Use Ruff for linting and formatting (configured in `pyproject.toml`)
- Maximum line length: 120 characters
- Use type hints where appropriate
- Follow the existing code style in the repository

**Linting Commands:**
```bash
# Check for issues
cd code/mcp
ruff check .

# Auto-fix issues
ruff check --fix .

# Format code
ruff format .
```

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

**Examples:**
```
feat: add cost anomaly detection model
fix: correct billing period calculation in silver layer
docs: update quickstart guide with new CLI commands
```

## Development Workflow

### For dbt Models

1. **Create or modify** your SQL model in `code/dbt/models/`
2. **Lint your SQL**:
   ```bash
   cd code/dbt
   sqlfluff lint models/your_model.sql
   sqlfluff fix models/your_model.sql  # Auto-fix issues
   ```
3. **Test locally**:
   ```bash
   dbt run --select your_model+
   dbt test --select your_model+
   ```
4. **Update documentation** in the corresponding `_models.yml` file

### For CLI (Go)

1. **Create or modify** your Go files in `code/cli/`
2. **Format your code**:
   ```bash
   cd code/cli
   gofmt -w .
   # Or use gofmt with specific files
   gofmt -w path/to/file.go
   ```
3. **Build and test**:
   ```bash
   go build
   go test ./...
   go test -v ./...  # With verbose output
   ```
4. **Run the CLI locally**:
   ```bash
   ./ecos --help
   ./ecos init --dry-run
   # Test specific commands
   ./ecos transform --help
   ```
5. **Follow Go conventions**:
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
   - Use meaningful variable and function names
   - Add comments for exported functions and types

### For Python Code

1. **Create or modify** your Python files in `code/mcp/`
2. **Lint and format**:
   ```bash
   cd code/mcp
   ruff check .
   ruff format .
   ```
3. **Run tests**:
   ```bash
   pytest
   pytest --cov  # With coverage
   ```

## Pull Request Process

1. **Ensure your code is tested** - All dbt models should compile and tests should pass
2. **Update documentation** - Add or update model descriptions in YAML files, update README if needed
3. **Follow the PR template** - Complete all relevant sections
4. **Request review** - Tag relevant maintainers if you know who should review
5. **Address feedback** - Respond to review comments and make requested changes

### PR Checklist

Before submitting, ensure:
- [ ] Code follows project style guidelines
- [ ] SQL files pass `sqlfluff lint`
- [ ] Python code passes `ruff check`
- [ ] All tests pass locally
- [ ] Documentation is updated
- [ ] Commit messages follow Conventional Commits format
- [ ] PR description is clear and references related issues

## Testing

### dbt Models

```bash
cd code/dbt

# Run all tests
dbt test

# Test specific model
dbt test --select your_model

# Run with verbose output
dbt test --verbose
```

### Python Code

```bash
cd code/mcp

# Run all tests
pytest

# Run with coverage
pytest --cov

# Run specific test file
pytest tests/test_unit.py
```

## Documentation

When adding new features or models:

1. **Update model documentation** in the corresponding `_models.yml` file
2. **Add column descriptions** for all new columns
3. **Update README.md** if adding new functionality
4. **Update web documentation** at [ecos-labs.io/docs](https://ecos-labs.io/docs) if it affects user-facing features

## Getting Help

- **Documentation**: [ecos-labs.io/docs](https://ecos-labs.io/docs)
- **Issues**: [GitHub Issues](https://github.com/ecos-labs/ecos-core/issues)
- **Search closed issues**: [Closed Issues](https://github.com/ecos-labs/ecos-core/issues?q=is%3Aissue+is%3Aclosed)

## Contributor Agreement

By submitting a pull request, you confirm that:

- This contribution is your original work, or you have the right to submit it
- You agree that this contribution will be licensed under the same license as the ecos project (Apache License 2.0)
- You have read and comply with the [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md)

This agreement is acknowledged by checking the boxes in the PR template.

---

Thank you for contributing to ecos! 🎉
