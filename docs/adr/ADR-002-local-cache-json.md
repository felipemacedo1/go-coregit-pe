# ADR-002: Use Local JSON Files for Cache Instead of External Database

## Status
Accepted

## Context
We need to implement caching and indexing for Git repository metadata to provide fast access for UI operations. We have several options:
1. Use an external database like SQLite, PostgreSQL, or MongoDB
2. Use local JSON/CSV files for simple key-value storage
3. Use in-memory caching only
4. Use a lightweight embedded database

## Decision
We will use local JSON files stored in `~/.gitmgr/cache/<repo-hash>/` for caching repository metadata.

## Rationale
- **No external dependencies**: Aligns with project constraint of using only standard library
- **Simplicity**: JSON is easy to read, write, and debug
- **Cross-platform**: Works consistently across Linux, macOS, and Windows
- **Human readable**: Cache files can be inspected and manually edited if needed
- **Lightweight**: No database setup or maintenance required
- **Atomic operations**: File operations can be made atomic with proper locking

## Implementation Details
- Cache structure: `~/.gitmgr/cache/<sha256-of-repo-path>/`
- Separate JSON files for different data types:
  - `branches.json` - Branch information
  - `remotes.json` - Remote repository information  
  - `commits.json` - Recent commit history
  - `status.json` - Last known repository status
- File-based locking to prevent concurrent access issues
- Polling-based file watching (mtime of `.git/HEAD`, `refs/`) for cache invalidation
- TTL-based expiration for cache entries

## Consequences
### Positive
- Zero external dependencies
- Simple implementation and debugging
- Cross-platform compatibility
- Human-readable cache files
- Easy backup and migration

### Negative
- Not suitable for very large repositories (performance may degrade)
- No complex querying capabilities
- Manual implementation of indexing and relationships
- File I/O overhead for frequent operations

## Alternatives Considered
- **SQLite**: Would provide better performance and querying but adds external dependency
- **In-memory only**: Would be fastest but loses data on restart
- **Binary formats**: Would be more compact but less debuggable

## Migration Path
If performance becomes an issue with large repositories, we can:
1. Add optional SQLite support while keeping JSON as default
2. Implement hybrid approach (JSON for small repos, SQLite for large ones)
3. Add compression for JSON files to reduce I/O overhead