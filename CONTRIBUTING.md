# Contributing to Dynapins Server

Thank you for your interest in contributing! 🎉

## Development Setup

### Prerequisites

- Go 1.25+ ([download](https://go.dev/dl/))
- Docker (optional, for containerized development)

### Getting Started

1. **Clone the repository:**

   ```bash
   git clone https://github.com/freecats/dynapins-server.git
   cd dynapins-server
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Generate test keys:**

   ```bash
   openssl genpkey -algorithm ED25519 -out private_key.pem
   ```

4. **Run tests:**

   ```bash
   make test
   ```

5. **Run the server:**

   ```bash
   export ALLOWED_DOMAINS="example.com"
   export PRIVATE_KEY_PEM="$(cat private_key.pem)"
   make run
   ```

## Making Changes

### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Run `make lint` to check for issues
- Ensure all tests pass: `make test`

### Testing

- Write tests for new features
- Maintain or improve code coverage
- Run `make test-coverage` to generate coverage report

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for RSA signatures
fix: handle timeout errors correctly
docs: update configuration examples
test: add integration tests for wildcards
```

## Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Commit your changes: `git commit -m 'feat: add amazing feature'`
6. Push to your fork: `git push origin feature/amazing-feature`
7. Open a Pull Request

### PR Checklist

- [ ] Tests pass locally
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG updated (for significant changes)

## Project Structure

```
pinning-server/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── cert/           # Certificate retrieval
│   ├── config/         # Configuration management
│   ├── domain/         # Domain validation
│   ├── logger/         # Structured logging
│   ├── server/         # HTTP server & handlers
│   └── sign/           # Cryptographic signing
├── Dockerfile          # Container build
├── Makefile            # Development commands
└── README.md           # User documentation
```

## Reporting Issues

Found a bug? Have a feature request?

1. Check if an issue already exists
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Go version, OS, and any relevant environment details

## Security

If you discover a security vulnerability, please email security@example.com instead of creating a public issue.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open a discussion or reach out to the maintainers.

Thank you for contributing! 🚀

