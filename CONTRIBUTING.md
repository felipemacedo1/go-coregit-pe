# Contributing to Go Core Git

Thank you for your interest in contributing to Go Core Git! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites
- Go 1.22 or later
- Git binary installed and available in PATH
- Make (for build automation)

### Getting Started
1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/felipemacedo1/go-coregit-pe.git
   cd go-coregit-pe
   ```
3. Build the project:
   ```bash
   make build
   ```
4. Run tests:
   ```bash
   make test
   ```

## Development Workflow

### Code Style
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused

### Testing
- Write tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Use table-driven tests where appropriate

### Commit Messages
We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(core): add merge operation support
fix(exec): handle timeout errors properly
docs(readme): update installation instructions
test(cache): add expiration test cases
```

### Branch Naming
- `feat/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `docs/documentation-update` - Documentation changes
- `chore/maintenance-task` - Maintenance tasks

## Architecture Guidelines

### Project Structure
```
/cmd/           - CLI applications
/internal/      - Private packages
/pkg/           - Public packages
/docs/          - Documentation
/scripts/       - Build and utility scripts
/testdata/      - Test fixtures
```

### Design Principles
1. **Stdlib Only**: Use only Go standard library, no external dependencies
2. **Security First**: Sanitize inputs, redact credentials, secure execution
3. **Cross-platform**: Support Linux, macOS, Windows
4. **Clean Architecture**: Clear separation of concerns, interfaces over implementations
5. **Git Native**: Execute native Git binary for full compatibility

### Interface Design
- Keep interfaces small and focused
- Use context.Context for cancellation and timeouts
- Return meaningful errors with context
- Follow Go idioms and conventions

## Testing Guidelines

### Test Categories
1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows

### Test Naming
```go
func TestFunctionName(t *testing.T)           // Basic test
func TestFunctionName_ErrorCase(t *testing.T) // Error case
func TestFunctionName_EdgeCase(t *testing.T)  // Edge case
```

### Test Structure
```go
func TestExample(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result, err := FunctionUnderTest(input)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %q, got %q", expected, result)
    }
}
```

## Documentation

### Code Documentation
- Document all exported functions, types, and constants
- Use complete sentences in comments
- Provide usage examples for complex functions

### ADRs (Architecture Decision Records)
- Document significant architectural decisions in `/docs/adr/`
- Follow the established ADR template
- Include context, decision, rationale, and consequences

### API Documentation
- Keep API specification up to date in `/docs/spec/`
- Include request/response examples
- Document error conditions and status codes

## Pull Request Process

1. **Create Feature Branch**: Branch from `main`
2. **Implement Changes**: Follow coding guidelines
3. **Add Tests**: Ensure good test coverage
4. **Update Documentation**: Update relevant docs
5. **Run Tests**: Ensure all tests pass
6. **Submit PR**: Create pull request with clear description

### PR Description Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
```

## Release Process

1. Update CHANGELOG.md
2. Update version in relevant files
3. Create release tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. Create GitHub release with changelog

## Getting Help

- Check existing issues and discussions
- Review documentation in `/docs/`
- Ask questions in GitHub issues
- Contact the maintainer: contato.dev.macedo@gmail.com
- Follow project conventions and patterns

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a positive environment

## Maintainer

**Felipe Macedo**
- GitHub: [github.com/felipemacedo1](https://github.com/felipemacedo1)
- LinkedIn: [linkedin.com/in/felipemacedo1](https://linkedin.com/in/felipemacedo1)
- Email: contato.dev.macedo@gmail.com

Thank you for contributing to Go Core Git!