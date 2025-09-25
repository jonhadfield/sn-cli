# Contributing to sn-cli

Thank you for your interest in contributing to sn-cli! We welcome contributions from the community.

## How to Contribute

### Reporting Issues

- Check if the issue already exists in [GitHub Issues](https://github.com/jonhadfield/sn-cli/issues)
- Include your OS, Go version, and sn-cli version
- Provide steps to reproduce the issue
- Include relevant error messages or logs

### Submitting Pull Requests

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/sn-cli.git
   cd sn-cli
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make Changes**
   - Follow existing code style and conventions
   - Add tests for new functionality
   - Update documentation if needed

4. **Test Your Changes**
   ```bash
   make test
   make lint
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "Brief description of changes"
   git push origin your-branch-name
   ```

6. **Open a Pull Request**
   - Provide a clear description of the changes
   - Reference any related issues
   - Ensure all checks pass

## Development Setup

### Prerequisites
- Go 1.25 or later
- Make

### Building
```bash
make build
```

### Running Tests
```bash
make test        # Run all tests
make cover       # Generate coverage report
```

### Code Quality
```bash
make lint        # Run linters
make fmt         # Format code
```

## Code Guidelines

- **Go Standards**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **Error Handling**: Always handle errors explicitly
- **Testing**: Write tests for new features and bug fixes
- **Comments**: Add comments for complex logic
- **Commits**: Use clear, descriptive commit messages

## Testing

- Unit tests should be included for new functionality
- Integration tests for API interactions are welcome
- Test with both Standard Notes cloud and self-hosted servers when possible

## Questions?

Feel free to open an issue for discussion or clarification about contributing.

## License

By contributing to sn-cli, you agree that your contributions will be licensed under the MIT License.
