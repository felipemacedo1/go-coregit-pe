package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/felipemacedo1/go-coregit-pe/pkg/core"
	"github.com/felipemacedo1/go-coregit-pe/pkg/core/execgit"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("gitmgr version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	case "repo":
		handleRepoCommand()
	case "clone":
		handleCloneCommand()
	case "status":
		handleStatusCommand()
	case "log":
		handleLogCommand()
	case "diff":
		handleDiffCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`gitmgr - Core Git Manager v%s

Usage:
  gitmgr <command> [options]

Commands:
  version         Show version information
  help            Show this help message
  repo open <path> Open an existing repository
  clone <url> [path] Clone a repository
  status [path]   Show repository status
  log [path]      Show commit history
  diff [path]     Show changes

More commands coming soon...
`, version)
}

func handleRepoCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: gitmgr repo <subcommand>\n")
		fmt.Fprintf(os.Stderr, "Subcommands: open\n")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "open":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: gitmgr repo open <path>\n")
			os.Exit(1)
		}
		path := os.Args[3]
		git := execgit.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		repo, err := git.Open(ctx, path)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Repository opened successfully:\n")
		fmt.Printf("  Path: %s\n", repo.Path)
		fmt.Printf("  Git Dir: %s\n", repo.GitDir)
		fmt.Printf("  Bare: %v\n", repo.IsBare)
		fmt.Printf("  Worktree: %v\n", repo.IsWorktree)
	default:
		fmt.Fprintf(os.Stderr, "Unknown repo subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleCloneCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: gitmgr clone <url> [path]\n")
		os.Exit(1)
	}

	url := os.Args[2]
	path := ""
	if len(os.Args) > 3 {
		path = os.Args[3]
	}

	// If no path provided, derive from URL
	if path == "" {
		// Simple extraction of repo name from URL
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			if strings.HasSuffix(lastPart, ".git") {
				path = lastPart[:len(lastPart)-4]
			} else {
				path = lastPart
			}
		}
		if path == "" {
			path = "repo"
		}
	}

	git := execgit.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	fmt.Printf("Cloning %s to %s...\n", url, path)
	repo, err := git.Clone(ctx, core.CloneOptions{
		URL:      url,
		Path:     path,
		Progress: true,
	})
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Repository cloned successfully to %s\n", repo.Path)
}

func handleStatusCommand() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	git := execgit.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	repo, err := git.Open(ctx, path)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	status, err := git.GetStatus(ctx, repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("On branch %s\n", status.Branch)
	if status.Upstream != "" {
		fmt.Printf("Your branch is ")
		switch {
		case status.Ahead > 0 && status.Behind > 0:
			fmt.Printf("ahead by %d and behind by %d commits", status.Ahead, status.Behind)
		case status.Ahead > 0:
			fmt.Printf("ahead by %d commits", status.Ahead)
		case status.Behind > 0:
			fmt.Printf("behind by %d commits", status.Behind)
		default:
			fmt.Printf("up to date")
		}
		fmt.Printf(" with '%s'\n", status.Upstream)
	}

	if status.Clean {
		fmt.Println("\nnothing to commit, working tree clean")
	} else {
		fmt.Printf("\nChanges in working directory:\n")
		for _, file := range status.Files {
			fmt.Printf("  %s %s\n", file.Status, file.Path)
		}
	}
}

func handleLogCommand() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	git := execgit.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	repo, err := git.Open(ctx, path)
	if err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	commits, err := git.Log(ctx, repo, "", 10, false)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, commit := range commits {
		fmt.Printf("commit %s\n", commit.Hash)
		fmt.Printf("Author: %s <%s>\n", commit.Author, commit.Email)
		fmt.Printf("Date:   %s\n", commit.Date.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Printf("\n    %s\n", commit.Subject)
		if commit.Body != "" {
			fmt.Printf("\n    %s\n", commit.Body)
		}
		fmt.Println()
	}
}

func handleDiffCommand() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	git := execgit.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	repo, err := git.Open(ctx, path)
	if err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	diff, err := git.Diff(ctx, repo, "", "", true)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("No changes")
	} else {
		fmt.Print(diff)
	}
}
