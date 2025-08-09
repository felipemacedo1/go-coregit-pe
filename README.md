# Go Core Git

A cross-platform Git manager with CLI and local API, built using only Go standard library.

## Features

- **Cross-platform**: Linux, macOS, Windows support
- **Security-first**: No credential logging, secure command execution
- **CLI interface**: `gitmgr` command with Git operations
- **Local HTTP API**: For automation and GUI integration
- **No external dependencies**: Uses only Go standard library
- **Native Git**: Executes native `git` binary for full compatibility

## Requirements

- Go 1.22 or later
- Git binary installed and available in PATH

## Installation

### From Source

```bash
git clone https://github.com/felipe-macedo/go-coregit-pe.git
cd go-coregit-pe
make build
```

The binary will be available at `bin/gitmgr`.

### Cross-compilation

```bash
make cross
```

Binaries for multiple platforms will be available in the `bin/` directory.

## Usage

### CLI Commands

```bash
# Show version
gitmgr version

# Show help
gitmgr help

# More commands coming soon...
```

### Development

```bash
# Run tests
make test

# Format code
make fmt

# Lint code
make lint

# Clean build artifacts
make clean
```

## Architecture

The project follows a clean architecture with clear separation of concerns:

- `cmd/gitmgr/`: CLI entry point
- `pkg/core/`: Public interfaces and types
- `pkg/core/execgit/`: Git binary execution implementation
- `internal/executil/`: Secure command execution utilities
- `internal/logging/`: Structured logging
- `docs/`: Documentation and ADRs

## Security

- All Git commands are executed with sanitized arguments
- Credentials are never logged or stored
- Uses Git's native credential helpers
- Minimal environment for command execution

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Status

🚧 **Work in Progress** - This project is under active development.

Current milestone: **Bootstrap** ✅
- [x] Project structure
- [x] Basic CLI with version command
- [x] Core interfaces and types
- [x] Secure Git executor
- [x] Structured logging
- [x] Documentation foundation

Next milestone: **Core Operations**
- [ ] Repository operations (open, clone, status)
- [ ] Branch operations
- [ ] Remote operations
- [ ] Basic CLI commands