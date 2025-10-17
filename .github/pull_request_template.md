## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Code style update (formatting, renaming)
- [ ] ♻️ Code refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] ✅ Test update
- [ ] 🔧 Build configuration change
- [ ] 🔒 Security fix

## Related Issue

<!-- Link to related issue(s) -->

Fixes #(issue number)
Closes #(issue number)
Related to #(issue number)

## Changes Made

<!-- List the main changes made in this PR -->

- 
- 
- 

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Configuration

- Go version:
- OS:
- Deployment method:

### Test Commands Run

```bash
# Example commands
make test
make lint
go test -race ./...
```

### Test Results

<!-- Paste test output or describe results -->

```
# Test output here
```

## Performance Impact

<!-- If applicable, describe any performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance degraded (explain why and if acceptable)

### Benchmark Results (if applicable)

```bash
# Before
BenchmarkFeature-8    1000000    1234 ns/op

# After
BenchmarkFeature-8    1500000     987 ns/op
```

## Breaking Changes

<!-- If this PR introduces breaking changes, describe them here -->

- [ ] This PR does not introduce breaking changes

**Breaking changes description:**

<!-- Describe what breaks and how users should adapt -->

## Documentation

<!-- Have you updated the documentation? -->

- [ ] README.md updated
- [ ] CHANGELOG.md updated
- [ ] API documentation updated
- [ ] Code comments updated
- [ ] No documentation needed

## Checklist

<!-- Mark completed items with an "x" -->

### Code Quality

- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or errors
- [ ] I have run `make lint` and fixed all issues

### Testing

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have run `make test` successfully
- [ ] I have run tests with race detector (`go test -race ./...`)
- [ ] I have tested the changes manually (if applicable)

### Security

- [ ] I have reviewed my code for security vulnerabilities
- [ ] No sensitive information (keys, passwords, tokens) is committed
- [ ] Dependencies are up to date and have no known vulnerabilities

### Commit Messages

- [ ] My commit messages follow the conventional commits format
- [ ] Commit messages are clear and descriptive

### Additional Checks

- [ ] I have checked my code for performance regressions
- [ ] I have updated error messages to be clear and actionable
- [ ] I have considered backward compatibility
- [ ] I have added appropriate logging

## Screenshots (if applicable)

<!-- Add screenshots to help explain your changes -->

## Additional Notes

<!-- Any additional information that reviewers should know -->

## Reviewers

<!-- Tag people you'd like to review this PR -->

@reviewer1 @reviewer2

---

**By submitting this pull request, I confirm that my contribution is made under the terms of the project's license.**
