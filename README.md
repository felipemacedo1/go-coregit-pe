# Go Core Git

A cross-platform Git manager with CLI and local API, built using only Go standard library.

## Features

- **Cross-platform**: Linux, macOS, Windows support
- **Security-first**: No credential logging, secure command execution
- **CLI interface**: `gitmgr` command with Git operations
- **HTTP API server**: `gitmgr-server` for automation and GUI integration
- **JSON-based cache**: Lightweight caching system for metadata
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

# Repository operations
gitmgr repo open /path/to/repo
gitmgr clone https://github.com/user/repo.git
gitmgr status

# View history and changes
gitmgr log
gitmgr diff

# More commands available - see gitmgr help
```

### HTTP API Server

```bash
# Start HTTP API server
gitmgr-server -addr=127.0.0.1:8080

# Use API endpoints
curl "http://127.0.0.1:8080/v1/status?path=/path/to/repo"
curl -X POST http://127.0.0.1:8080/v1/clone \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/user/repo.git","path":"/local/path"}'

# See docs/spec/api-spec.md for full API documentation
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

Current milestone: **Core Operations** ✅
- [x] Repository operations (open, clone, status, init, discover, config)
- [x] Remote operations (list, add, remove, set-url)
- [x] Sync operations (fetch, pull, push) with error handling
- [x] Branch operations (create, delete, checkout, list)
- [x] Inspection operations (log, diff)
- [x] CLI commands (repo, clone, status, log, diff)
- [x] HTTP API server with RESTful endpoints
- [x] JSON-based cache system with TTL
- [x] Comprehensive test coverage and documentation

Next milestone: **Advanced Operations**
- [ ] Merge/Rebase/Cherry-pick operations
- [ ] Tag operations
- [ ] Stash operations
- [ ] Worktree operations
- [ ] Submodule and LFS support