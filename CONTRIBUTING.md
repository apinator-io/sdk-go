# Contributing to Apinator Go SDK

Thank you for your interest in contributing to the Apinator Go SDK.

## Getting Started

1. Fork the repository: [github.com/apinator-io/sdk-go](https://github.com/apinator-io/sdk-go)
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/sdk-go.git
   cd sdk-go
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feat/your-feature
   ```

## Development Requirements

- Go 1.21 or later
- No external dependencies — stdlib only

## Running Tests

```bash
go test -race -count=1 -v ./...
```

## Code Quality

Before submitting a pull request, ensure your code passes:

```bash
# Run tests with race detection
go test -race -count=1 -v ./...

# Run the Go vet static analysis tool
go vet ./...
```

## Guidelines

- **Zero external dependencies.** This SDK uses the Go standard library only. Do not add third-party modules.
- **Test coverage.** All new functionality must include tests. Aim for 85%+ coverage.
- **Strict types.** Use concrete types where possible. Avoid `interface{}` unless necessary.
- **Doc comments.** Every exported function, type, and method must have a Go doc comment.
- **Error handling.** Return typed errors (`ApiError`, `AuthenticationError`, `ValidationError`) where appropriate.

## Commit Messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add connection timeout option
fix: correct HMAC signature for empty body
chore: update Go version in CI
docs: add webhook verification example
test: add coverage for auth edge cases
```

Breaking changes use `!`:

```
feat!: rename RealtimeClient to Client
```

## Pull Requests

1. Keep PRs focused on a single change.
2. Update tests and documentation as needed.
3. Reference any related issues in the PR description.
4. Ensure all CI checks pass before requesting review.

## Releasing

Releases are managed by maintainers using [release-please](https://github.com/googleapis/release-please). Conventional commit messages drive automatic changelog generation and version bumps.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
