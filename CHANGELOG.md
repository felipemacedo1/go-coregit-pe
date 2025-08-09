# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- HTTP API server with RESTful endpoints
- JSON-based cache system with TTL support
- Log and diff operations with CLI commands
- Comprehensive test coverage
- API specification documentation
- Demo and test scripts

### Changed
- Expanded CLI with repository operations
- Enhanced error handling with user-friendly messages
- Updated documentation with current features

## [0.3.0] - 2025-08-09

### Added
- HTTP API server (`gitmgr-server`) with RESTful endpoints
- API endpoints for repository, sync, and inspection operations
- Raw command execution endpoint for flexibility
- Graceful server shutdown handling
- Comprehensive API specification documentation
- API test script for validation

### Changed
- Updated Makefile to build both CLI and server binaries
- Enhanced README with API usage examples

## [0.2.0] - 2025-08-09

### Added
- Log operation with custom format parsing
- Diff operation with stat support
- Log and diff CLI commands with proper formatting
- JSON-based cache system in `pkg/index`
- Cache support for branches, remotes, status, and commits
- TTL-based cache expiration
- Comprehensive cache tests
- ADR-002 documenting local JSON cache decision

### Changed
- Updated README with current progress and CLI examples

## [0.1.0] - 2025-08-09

### Added
- Core Git operations implementation
- Repository operations: Init, Discover, GetConfig, SetConfig
- Remote operations: ListRemotes, AddRemote, RemoveRemote, SetRemoteURL
- Sync operations: Fetch, Pull, Push with error handling
- Branch operations: CreateBranch, DeleteBranch, Checkout, ListBranches
- CLI commands: repo open, clone, status
- Comprehensive error handling with user-friendly messages
- Basic test coverage for executil and execgit packages
- Demo script showcasing functionality
- URL sanitization for secure logging

### Changed
- Expanded CLI from basic version command to full operations

## [0.0.1] - 2025-08-09

### Added
- Initial project bootstrap
- Go module and Makefile for build automation
- CLI entry point with version command
- Core interfaces and types for Git operations
- Secure Git command executor with sanitization
- Structured logging using stdlib only
- Initial ExecGit implementation with Open, Clone, GetStatus
- Project structure following clean architecture
- ADR-001 documenting decision to use Git binary
- Comprehensive documentation foundation

### Security
- Implemented argument sanitization to prevent injection
- Added credential redaction in logs and output
- Secure environment setup for Git command execution