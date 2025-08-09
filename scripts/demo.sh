#!/bin/bash
set -e

echo "=== Go Core Git Demo ==="
echo

# Build the binary
echo "Building gitmgr..."
make build
echo

# Show version
echo "Version:"
./bin/gitmgr version
echo

# Show help
echo "Help:"
./bin/gitmgr help
echo

# Test repo open on current directory
echo "Opening current repository:"
./bin/gitmgr repo open .
echo

# Test status
echo "Repository status:"
./bin/gitmgr status
echo

echo "Demo completed successfully!"
echo
echo "Try these commands:"
echo "  ./bin/gitmgr repo open <path>    # Open a repository"
echo "  ./bin/gitmgr clone <url> [path]  # Clone a repository"
echo "  ./bin/gitmgr status [path]       # Show repository status"