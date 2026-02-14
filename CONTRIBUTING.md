# Contributing

Thank you for your interest in contributing to scdl.

## Reporting Issues

If you find a bug or have a feature request, please open an issue and include:

- A clear description of the problem
- Steps to reproduce (if applicable)
- Expected vs actual behavior
- Your environment (OS, Go version)

Keep issues focused and reproducible.

## Development Setup

Requirements:

- Go (latest stable version recommended)

Clone the repository:

```bash
git clone git@github.com:hellsontime/scdl.git
cd scdl
go mod tidy
```

Build:

```bash
go build ./...
```

## Code Style

- Follow standard Go formatting (`gofmt`)
- Use clear, explicit error handling
- Avoid unnecessary dependencies
- Keep functions small and focused
- Prefer simple and readable implementations

Before submitting:

```bash
make check
make check-coverage
```

## Commit Messages

This project follows Conventional Commits.

Examples:

- `feat: add progressive stream detection`
- `fix: handle expired signed url`
- `docs: update README usage section`
- `refactor: simplify transcoding resolver`

## Pull Requests

- Keep PRs small and focused
- Describe what changed and why
- Reference related issues if applicable
- Ensure the project builds successfully

## Code of Conduct

By participating in this project, you agree to abide by the Code of Conduct.
