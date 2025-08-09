# HTTP API Specification

## Overview
The gitmgr HTTP API provides RESTful endpoints for Git operations. All endpoints return JSON responses with a consistent structure.

## Base URL
```
http://127.0.0.1:8080
```

## Response Format
All responses follow this structure:
```json
{
  "success": true|false,
  "data": <response_data>,
  "error": "<error_message>"
}
```

## Authentication
Currently no authentication is required. The API is designed for local use only.

## Endpoints

### Health Check
```
GET /health
```
Returns server health status.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "time": "2025-01-01T12:00:00Z"
  }
}
```

### Repository Info
```
GET /v1/repo?path=<repo_path>
```
Get repository information.

**Parameters:**
- `path` (required): Repository path

**Response:**
```json
{
  "success": true,
  "data": {
    "path": "/path/to/repo",
    "workDir": "/path/to/repo",
    "gitDir": "/path/to/repo/.git",
    "isBare": false,
    "isWorktree": true
  }
}
```

### Clone Repository
```
POST /v1/clone
```
Clone a repository.

**Request Body:**
```json
{
  "url": "https://github.com/user/repo.git",
  "path": "/local/path",
  "branch": "main",
  "depth": 1,
  "sparse": ["src/", "docs/"],
  "recursive": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "path": "/local/path",
    "workDir": "/local/path",
    "gitDir": "/local/path/.git",
    "isBare": false,
    "isWorktree": true
  }
}
```

### Repository Status
```
GET /v1/status?path=<repo_path>
```
Get repository status.

**Parameters:**
- `path` (required): Repository path

**Response:**
```json
{
  "success": true,
  "data": {
    "branch": "main",
    "upstream": "origin/main",
    "ahead": 0,
    "behind": 0,
    "files": [
      {
        "path": "file.txt",
        "status": " M",
        "staged": false,
        "modified": true
      }
    ],
    "clean": false
  }
}
```

### Commit Log
```
GET /v1/log?path=<repo_path>&max=<count>
```
Get commit history.

**Parameters:**
- `path` (required): Repository path
- `max` (optional): Maximum number of commits (default: 10)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "hash": "abc123...",
      "shortHash": "abc123",
      "author": "John Doe",
      "email": "john@example.com",
      "date": "2025-01-01T12:00:00Z",
      "subject": "feat: add new feature",
      "body": "Detailed description..."
    }
  ]
}
```

### Diff
```
GET /v1/diff?path=<repo_path>&base=<base>&head=<head>&stat=<true|false>
```
Get diff between commits or working directory.

**Parameters:**
- `path` (required): Repository path
- `base` (optional): Base commit/branch
- `head` (optional): Head commit/branch
- `stat` (optional): Show only statistics (default: false)

**Response:**
```json
{
  "success": true,
  "data": {
    "diff": "diff --git a/file.txt b/file.txt\n..."
  }
}
```

### Fetch
```
POST /v1/fetch
```
Fetch from remote repository.

**Request Body:**
```json
{
  "path": "/repo/path",
  "remote": "origin",
  "prune": true,
  "tags": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Fetch completed successfully"
  }
}
```

### Pull
```
POST /v1/pull
```
Pull from remote repository.

**Request Body:**
```json
{
  "path": "/repo/path",
  "remote": "origin",
  "branch": "main",
  "rebase": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Pull completed successfully"
  }
}
```

### Push
```
POST /v1/push
```
Push to remote repository.

**Request Body:**
```json
{
  "path": "/repo/path",
  "remote": "origin",
  "branch": "main",
  "force": false,
  "tags": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Push completed successfully"
  }
}
```

### Raw Command
```
POST /v1/raw
```
Execute raw git command.

**Request Body:**
```json
{
  "path": "/repo/path",
  "args": ["status", "--porcelain"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "exitCode": 0,
    "stdout": "output...",
    "stderr": "",
    "duration": 1000000000
  }
}
```

## Error Responses
Error responses include an error message:
```json
{
  "success": false,
  "error": "Repository not found"
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `405` - Method Not Allowed
- `500` - Internal Server Error

## Usage Examples

### Using curl
```bash
# Get repository status
curl "http://127.0.0.1:8080/v1/status?path=/path/to/repo"

# Clone repository
curl -X POST http://127.0.0.1:8080/v1/clone \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/user/repo.git","path":"/local/path"}'

# Execute raw command
curl -X POST http://127.0.0.1:8080/v1/raw \
  -H "Content-Type: application/json" \
  -d '{"path":"/repo/path","args":["log","--oneline","-5"]}'
```