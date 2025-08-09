# Go Core Git

[![CI](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/ci.yml/badge.svg)](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/ci.yml)
[![Release](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/release.yml/badge.svg)](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/release.yml)
[![CodeQL](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/codeql.yml/badge.svg)](https://github.com/felipemacedo1/go-coregit-pe/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/felipemacedo1/go-coregit-pe)](https://goreportcard.com/report/github.com/felipemacedo1/go-coregit-pe)

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

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/felipemacedo1/go-coregit-pe/releases):

```bash
# Linux/macOS
curl -L https://github.com/felipemacedo1/go-coregit-pe/releases/latest/download/gitmgr-linux-amd64 -o gitmgr
chmod +x gitmgr
sudo mv gitmgr /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/felipemacedo1/go-coregit-pe.git
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

Custom License - Free for personal and educational use. Commercial and enterprise use requires authorization.
See [LICENSE](LICENSE) for details.

For commercial licensing information, see [COMMERCIAL-LICENSE.md](docs/COMMERCIAL-LICENSE.md)
Contact: contato.dev.macedo@gmail.com

## Author

**Felipe Macedo**
- Portfolio: [felipemacedo1.github.io](https://felipemacedo1.github.io/)
- LinkedIn: [linkedin.com/in/felipemacedo1](https://linkedin.com/in/felipemacedo1)

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