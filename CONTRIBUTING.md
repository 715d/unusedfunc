# Contributing to unusedfunc

Thank you for contributing to unusedfunc! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.25 or later
- golangci-lint (for linting)
- goimports (for formatting)

Install development tools:
```bash
make tools
```

### Building

Build the binary:
```bash
make build
# Output: build/unusedfunc
```

### Testing

Run all tests:
```bash
make test              # Run all tests with race detector and coverage
```

### Code Quality

Format code:
```bash
make fmt
```

Run linters:
```bash
make lint
```

## Code Standards

### Go Style

- Follow standard Go conventions
- Use early returns to reduce nesting
- Comments should explain "why", not "what"
- End comments with periods
- Use `any` instead of `interface{}`
- Error messages don't start with "failed to"

### Testing

- Write tests for new functionality
- Tests should verify actual behavior, not mocks
- Use testify for assertions (already in go.mod)
- Add edge cases to `testdata/` directory with `expected.yaml`

## Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Run tests and linters: `make test && make lint`
5. Commit with clear messages
6. Push to your fork
7. Open a pull request

## Release Process

Releases are automated via GitHub Actions and GoReleaser.

### Creating a Release

1. **Ensure main branch is stable**
   ```bash
   git checkout main
   git pull origin main
   make test && make lint
   ```

2. **Create and push a version tag**
   ```bash
   # Create an annotated tag (format: vMAJOR.MINOR.PATCH)
   git tag -a v1.0.1 -m "Release v1.0.1"

   # Push the tag to GitHub
   git push origin v1.0.1
   ```

3. **GitHub Actions automatically:**
   - Runs full test suite
   - Runs golangci-lint
   - Builds binaries for:
     - macOS (Intel and Apple Silicon)
     - Linux (amd64 and arm64)
     - Windows (amd64)
   - Creates GitHub release with auto-generated changelog
   - Updates Homebrew formula in `Formula/unusedfunc.rb`

4. **Verify the release**
   - Check GitHub releases page: https://github.com/715d/unusedfunc/releases
   - Verify all binaries are present
   - Test Homebrew installation:
     ```bash
     brew install 715d/unusedfunc/unusedfunc
     unusedfunc --version
     ```

### Testing Release Locally

Test the release process without publishing:

```bash
# Install GoReleaser if not already installed
go install github.com/goreleaser/goreleaser/v2@latest

# Run snapshot build (no publish, no tag validation)
goreleaser release --snapshot --clean

# Binaries will be in dist/ directory
ls -la dist/
```

This is useful for:
- Validating `.goreleaser.yaml` configuration
- Testing cross-compilation before tagging
- Verifying archive contents and structure

### Release Versioning

Follow semantic versioning (SemVer):
- **MAJOR** (v2.0.0): Breaking changes, incompatible API changes
- **MINOR** (v1.1.0): New features, backwards compatible
- **PATCH** (v1.0.1): Bug fixes, backwards compatible

### Changelog

The changelog is auto-generated from commit messages. For better changelogs, use conventional commit format:

- `feat: add support for X` → Features section
- `fix: resolve issue with Y` → Bug Fixes section
- `perf: optimize Z` → Performance section
- `docs: update README` → Excluded from changelog
- `test: add tests for W` → Excluded from changelog

### Rollback a Release

If a release has issues:

1. Delete the GitHub release (via web UI)
2. Delete the tag:
   ```bash
   git tag -d v1.0.1                    # Delete local
   git push origin --delete v1.0.1      # Delete remote
   ```
3. Fix the issue on main branch
4. Create a new patch release

### Release Artifacts

Each release includes:
- **Binaries:** Pre-built for all supported platforms
- **Archives:** tar.gz (macOS/Linux), zip (Windows)
- **Checksums:** SHA256 checksums in `checksums.txt`
- **Source code:** Automatic GitHub archives
- **Homebrew formula:** Auto-updated in repository

## Questions?

- Open an issue for bugs or feature requests
- Check `docs/architecture.md` for technical details
- Review `docs/workflows.md` for development workflows
