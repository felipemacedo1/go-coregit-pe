# ADR-001: Use Git Binary Instead of Reimplementing Git Protocol

## Status
Accepted

## Context
We need to implement Git operations for our Core Git manager. We have two main options:
1. Reimplement Git protocol and operations from scratch
2. Execute the native `git` binary and parse its output

## Decision
We will execute the native `git` binary for all Git operations.

## Rationale
- **Reliability**: Git binary is battle-tested and handles edge cases we might miss
- **Compatibility**: Full compatibility with Git ecosystem (hooks, configs, etc.)
- **Security**: Git handles authentication, SSL, and security properly
- **Maintenance**: No need to keep up with Git protocol changes
- **Features**: Immediate access to all Git features (LFS, submodules, etc.)

## Consequences
### Positive
- Faster development time
- Full Git compatibility
- Reduced maintenance burden
- Access to all Git features

### Negative
- Dependency on Git binary being installed
- Need to parse Git output formats
- Potential performance overhead from process spawning

## Implementation Notes
- Use `exec.CommandContext` for timeout control
- Sanitize arguments to prevent injection
- Parse structured output where possible (porcelain formats)
- Map common error patterns to user-friendly messages