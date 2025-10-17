# Contributing to Dynapins Server

Thank you for your interest in contributing to Dynapins! 🎉

We welcome contributions from the community and are grateful for your support.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Style Guidelines](#style-guidelines)
- [Community](#community)

## 🤝 Code of Conduct

This project and everyone participating in it is governed by our commitment to fostering an open and welcoming environment. Please be respectful and constructive in all interactions.

## 🚀 How Can I Contribute?

### Reporting Bugs

Found a bug? Please help us fix it!

1. **Check existing issues** - Someone might have already reported it
2. **Create a new issue** - Use our [bug report template](https://github.com/Free-cat/dynapins-server/issues/new?template=bug_report.yml)
3. **Provide details**:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Go version, OS, and environment details
   - Error messages or logs

### Suggesting Enhancements

Have an idea for improvement?

1. **Check existing issues** - Your idea might already be discussed
2. **Open a feature request** - Use our [feature request template](https://github.com/Free-cat/dynapins-server/issues/new?template=feature_request.yml)
3. **Describe your suggestion**:
   - Clear use case
   - Expected behavior
   - Why this would be useful
   - Possible implementation approach

### Contributing Code

Ready to write code? Great!

1. **Find an issue** - Look for issues labeled `good first issue` or `help wanted`
2. **Comment on the issue** - Let us know you're working on it
3. **Fork and create a branch** - Follow our [development setup](#development-setup)
4. **Make your changes** - Follow our [style guidelines](#style-guidelines)
5. **Submit a Pull Request** - Follow our [PR guidelines](#submitting-changes)

### Improving Documentation

Documentation is crucial! You can help by:

- Fixing typos or clarifying existing docs
- Adding examples or use cases
- Translating documentation
- Writing tutorials or blog posts

## 🛠️ Development Setup

### Prerequisites

- **Go 1.25+** - [Download](https://go.dev/dl/)
- **Docker** (optional) - For testing containerized builds
- **Git** - For version control

### Setup Steps

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/dynapins-server.git
   cd dynapins-server
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/Free-cat/dynapins-server.git
   ```

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Generate test keys:**
   ```bash
   openssl genpkey -algorithm ED25519 -out private_key.pem
   openssl pkey -in private_key.pem -pubout -out public_key.pem
   ```

6. **Run tests:**
   ```bash
   make test
   ```

7. **Start the server:**
   ```bash
   export ALLOWED_DOMAINS="example.com"
   export PRIVATE_KEY_PEM="$(cat private_key.pem)"
   make run
   ```

## ✏️ Making Changes

### Creating a Branch

Always create a new branch for your changes:

```bash
git checkout -b feature/amazing-feature
# or
git checkout -b fix/bug-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding or updating tests

### Writing Code

Follow these guidelines:

1. **Write clean, readable code**
2. **Add tests** for new functionality
3. **Update documentation** if needed
4. **Keep changes focused** - One feature/fix per PR
5. **Follow Go conventions** - Use `gofmt` and `golint`

### Testing

Ensure all tests pass before submitting:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific tests
go test ./internal/domain -v

# Run linters
make lint
```

## 📤 Submitting Changes

### Pull Request Process

1. **Update your branch** with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Commit your changes** with clear messages:
   ```bash
   git commit -m "feat: add certificate caching"
   ```

   Commit message format:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation
   - `test:` - Tests
   - `refactor:` - Code refactoring
   - `perf:` - Performance improvement
   - `chore:` - Maintenance

3. **Push to your fork:**
   ```bash
   git push origin feature/amazing-feature
   ```

4. **Create a Pull Request** on GitHub

### PR Checklist

Before submitting, make sure:

- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] PR description explains what and why
- [ ] Screenshots/logs included (if UI/output changed)

### PR Review Process

1. **Automated checks** run automatically
2. **Maintainer review** - Usually within 2-3 days
3. **Address feedback** - Make requested changes
4. **Approval & merge** - Once approved, we'll merge your PR

## 📐 Style Guidelines

### Go Code Style

Follow standard Go conventions:

```go
// Good: Clear function names and comments
// GetCertificates retrieves TLS certificates for the specified domain.
func GetCertificates(domain string) ([]*x509.Certificate, error) {
    // Implementation
}

// Good: Error handling
if err != nil {
    return nil, fmt.Errorf("failed to retrieve certificates: %w", err)
}

// Good: Table-driven tests
func TestDomainValidator(t *testing.T) {
    tests := []struct {
        name    string
        domain  string
        wantErr bool
    }{
        // Test cases
    }
}
```

### Documentation Style

- Use clear, concise language
- Include code examples
- Explain the "why", not just the "how"
- Keep lines under 100 characters
- Use proper Markdown formatting

### Commit Messages

Good commit messages:
```
✅ feat: add support for RSA signatures
✅ fix: handle timeout errors correctly
✅ docs: update configuration examples
✅ test: add integration tests for wildcards
```

Bad commit messages:
```
❌ Update stuff
❌ Fix bug
❌ WIP
```

## 💬 Community

### Getting Help

- **New Issue** - Report bugs or request features: [Create Issue](https://github.com/Free-cat/dynapins-server/issues/new/choose)
- **Issues** - Browse existing issues: [GitHub Issues](https://github.com/Free-cat/dynapins-server/issues)
- **Discussions** - Ask questions: [GitHub Discussions](https://github.com/Free-cat/dynapins-server/discussions)
- **Documentation** - Check our [README](README.md)

### Staying Updated

- **Watch** the repository for updates
- **Star** the project if you find it useful
- **Follow** releases for new versions

## 🏆 Recognition

Contributors are recognized in:
- Release notes
- GitHub contributors page
- Project README (for significant contributions)

## ❓ Questions?

If you have questions:

1. Check existing [Issues](https://github.com/Free-cat/dynapins-server/issues)
2. Search [Discussions](https://github.com/Free-cat/dynapins-server/discussions)
3. Create a new issue or discussion

## 📄 License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to Dynapins! Every contribution, no matter how small, makes a difference. 🙏
