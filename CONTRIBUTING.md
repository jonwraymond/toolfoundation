# Contributing to toolfoundation

Thank you for your interest in contributing to toolfoundation.

## Development Setup

### Prerequisites

- Go 1.24 or later
- golangci-lint (for linting)
- gosec (for security scanning)

### Clone and Build

```bash
git clone https://github.com/jonwraymond/toolfoundation.git
cd toolfoundation
go mod download
go build ./...
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Race Detection

```bash
go test -race ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Fuzz Tests

Fuzz tests are available for version parsing, tag normalization, and schema validation:

```bash
# Run a specific fuzz test for 30 seconds
go test -fuzz=FuzzParse$ -fuzztime=30s ./version/...
go test -fuzz=FuzzNormalizeTags$ -fuzztime=30s ./model/...

# Run all fuzz tests with seed corpus only
go test ./...
```

## Code Quality

### Linting

We use golangci-lint with the configuration in `.golangci.yml`:

```bash
golangci-lint run
```

### Security Scanning

```bash
gosec ./...
```

### Vulnerability Checking

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/). Commit messages are validated by commitlint in CI.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(adapter): add Anthropic cache control support

fix(model): handle empty namespace in tool ID parsing

docs(readme): add installation instructions

test(version): add fuzz tests for constraint parsing
```

## Pull Request Process

1. **Fork the repository** and create your branch from `main`.

2. **Write tests** for any new functionality. Maintain or improve test coverage.

3. **Run the full test suite** to ensure your changes don't break existing functionality:
   ```bash
   go test -race ./...
   golangci-lint run
   ```

4. **Update documentation** if you're changing public APIs or behavior.

5. **Use conventional commit messages** for your commits.

6. **Submit your PR** with a clear description of the changes.

## Code Style

- Follow standard Go conventions and `gofmt` formatting
- Use meaningful variable and function names
- Add doc comments to all exported types and functions
- Keep functions focused and reasonably sized
- Prefer returning errors over panicking

## Package Guidelines

### model

- All tool definitions must be MCP-spec compliant
- Schema validation must not perform network I/O
- Tag normalization must be deterministic

### adapter

- Conversions must be pure (no side effects)
- Feature loss must be reported via warnings, not errors
- New adapters must implement the full `Adapter` interface

### version

- Follow semantic versioning (SemVer 2.0.0)
- Version parsing must handle both `v1.0.0` and `1.0.0` formats

## Questions?

Open an issue for questions or discussions about contributions.
