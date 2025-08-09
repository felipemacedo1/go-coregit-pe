package execgit

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/felipe-macedo/go-coregit-pe/internal/executil"
	"github.com/felipe-macedo/go-coregit-pe/internal/logging"
	"github.com/felipe-macedo/go-coregit-pe/pkg/core"
)

// ExecGit implements CoreGit interface using git binary
type ExecGit struct {
	executor *executil.GitExecutor
	logger   *logging.Logger
}

// New creates a new ExecGit instance
func New() *ExecGit {
	return &ExecGit{
		executor: executil.NewGitExecutor(),
		logger:   logging.NewLogger(nil, false),
	}
}

// Open opens an existing repository
func (e *ExecGit) Open(ctx context.Context, path string) (*core.Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Check if it's a git repository
	result, err := e.executor.Run(ctx, absPath, []string{"rev-parse", "--git-dir"})
	if err != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("not a git repository: %s", absPath)
	}

	gitDir := strings.TrimSpace(result.Stdout)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(absPath, gitDir)
	}

	// Check if bare repository
	result, err = e.executor.Run(ctx, absPath, []string{"rev-parse", "--is-bare-repository"})
	if err != nil {
		return nil, fmt.Errorf("failed to check if bare repository: %w", err)
	}
	
	isBare := strings.TrimSpace(result.Stdout) == "true"

	// Check if worktree
	result, err = e.executor.Run(ctx, absPath, []string{"rev-parse", "--is-inside-work-tree"})
	isWorktree := err == nil && result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "true"

	repo := &core.Repo{
		Path:       absPath,
		WorkDir:    absPath,
		GitDir:     gitDir,
		IsBare:     isBare,
		IsWorktree: isWorktree,
	}

	e.logger.Info("Opened repository", map[string]interface{}{
		"path":    absPath,
		"bare":    isBare,
		"worktree": isWorktree,
	})

	return repo, nil
}

// Clone clones a repository
func (e *ExecGit) Clone(ctx context.Context, opts core.CloneOptions) (*core.Repo, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("clone URL is required")
	}
	if opts.Path == "" {
		return nil, fmt.Errorf("clone path is required")
	}

	args := []string{"clone"}
	
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}
	if opts.Depth > 0 {
		args = append(args, "--depth", strconv.Itoa(opts.Depth))
	}
	if opts.Bare {
		args = append(args, "--bare")
	}
	if opts.Mirror {
		args = append(args, "--mirror")
	}
	if opts.Recursive {
		args = append(args, "--recursive")
	}
	if opts.Progress {
		args = append(args, "--progress")
	}

	args = append(args, opts.URL, opts.Path)

	e.logger.Info("Cloning repository", map[string]interface{}{
		"url":    sanitizeURL(opts.URL),
		"path":   opts.Path,
		"branch": opts.Branch,
		"depth":  opts.Depth,
	})

	// Clone from parent directory
	parentDir := filepath.Dir(opts.Path)
	result, err := e.executor.Run(ctx, parentDir, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute clone: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("clone failed: %s", result.Stderr)
	}

	// Open the cloned repository
	return e.Open(ctx, opts.Path)
}

// GetStatus gets repository status
func (e *ExecGit) GetStatus(ctx context.Context, repo *core.Repo) (*core.RepoStatus, error) {
	// Get current branch
	result, err := e.executor.Run(ctx, repo.Path, []string{"branch", "--show-current"})
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	
	branch := strings.TrimSpace(result.Stdout)

	// Get upstream info
	upstream := ""
	ahead := 0
	behind := 0
	
	if branch != "" {
		result, err = e.executor.Run(ctx, repo.Path, []string{"rev-parse", "--abbrev-ref", branch + "@{upstream}"})
		if err == nil && result.ExitCode == 0 {
			upstream = strings.TrimSpace(result.Stdout)
			
			// Get ahead/behind counts
			result, err = e.executor.Run(ctx, repo.Path, []string{"rev-list", "--count", "--left-right", branch + "..." + upstream})
			if err == nil && result.ExitCode == 0 {
				counts := strings.Fields(strings.TrimSpace(result.Stdout))
				if len(counts) == 2 {
					ahead, _ = strconv.Atoi(counts[0])
					behind, _ = strconv.Atoi(counts[1])
				}
			}
		}
	}

	// Get file status
	result, err = e.executor.Run(ctx, repo.Path, []string{"status", "--porcelain"})
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []core.FileStatus
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}
		if len(line) < 3 {
			continue
		}
		
		status := line[:2]
		path := line[3:]
		
		files = append(files, core.FileStatus{
			Path:     path,
			Status:   status,
			Staged:   status[0] != ' ' && status[0] != '?',
			Modified: status[1] != ' ',
		})
	}

	return &core.RepoStatus{
		Branch:   branch,
		Upstream: upstream,
		Ahead:    ahead,
		Behind:   behind,
		Files:    files,
		Clean:    len(files) == 0,
	}, nil
}

// RunRaw executes a raw git command
func (e *ExecGit) RunRaw(ctx context.Context, repo *core.Repo, args []string) (*core.ExecResult, error) {
	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return nil, err
	}

	return &core.ExecResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: result.Duration,
	}, nil
}

// sanitizeURL removes credentials from URL for logging
func sanitizeURL(url string) string {
	if strings.Contains(url, "@") && (strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")) {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			protocol := strings.Split(parts[0], "://")[0]
			return protocol + "://***@" + strings.Join(parts[1:], "@")
		}
	}
	return url
}

// Placeholder implementations for other interface methods
// These will be implemented in subsequent commits

func (e *ExecGit) Init(ctx context.Context, path string, bare bool) (*core.Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}
	args = append(args, absPath)

	e.logger.Info("Initializing repository", map[string]interface{}{
		"path": absPath,
		"bare": bare,
	})

	// Init from parent directory
	parentDir := filepath.Dir(absPath)
	result, err := e.executor.Run(ctx, parentDir, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute init: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("init failed: %s", result.Stderr)
	}

	// Open the initialized repository
	return e.Open(ctx, absPath)
}

func (e *ExecGit) Discover(ctx context.Context, path string) (*core.Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Find the git repository root
	result, err := e.executor.Run(ctx, absPath, []string{"rev-parse", "--show-toplevel"})
	if err != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("no git repository found at %s", absPath)
	}

	repoRoot := strings.TrimSpace(result.Stdout)
	return e.Open(ctx, repoRoot)
}

func (e *ExecGit) GetConfig(ctx context.Context, repo *core.Repo, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("config key is required")
	}

	result, err := e.executor.Run(ctx, repo.Path, []string{"config", "--get", key})
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	return strings.TrimSpace(result.Stdout), nil
}

func (e *ExecGit) SetConfig(ctx context.Context, repo *core.Repo, key, value string, global bool) error {
	if key == "" {
		return fmt.Errorf("config key is required")
	}

	args := []string{"config"}
	if global {
		args = append(args, "--global")
	}
	args = append(args, key, value)

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to set config %s: %s", key, result.Stderr)
	}

	e.logger.Info("Config updated", map[string]interface{}{
		"key":    key,
		"global": global,
	})

	return nil
}

func (e *ExecGit) ListRemotes(ctx context.Context, repo *core.Repo) ([]core.RemoteInfo, error) {
	result, err := e.executor.Run(ctx, repo.Path, []string{"remote", "-v"})
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to list remotes: %s", result.Stderr)
	}

	remotes := make(map[string]*core.RemoteInfo)
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		name := parts[0]
		url := parts[1]
		type_ := strings.Trim(parts[2], "()")

		if remote, exists := remotes[name]; exists {
			if type_ == "fetch" {
				remote.FetchURL = url
			} else if type_ == "push" {
				remote.PushURL = url
			}
		} else {
			remote := &core.RemoteInfo{
				Name: name,
				URL:  url,
			}
			if type_ == "fetch" {
				remote.FetchURL = url
			} else if type_ == "push" {
				remote.PushURL = url
			}
			remotes[name] = remote
		}
	}

	var result_list []core.RemoteInfo
	for _, remote := range remotes {
		result_list = append(result_list, *remote)
	}

	return result_list, nil
}

func (e *ExecGit) AddRemote(ctx context.Context, repo *core.Repo, name, url string) error {
	if name == "" {
		return fmt.Errorf("remote name is required")
	}
	if url == "" {
		return fmt.Errorf("remote URL is required")
	}

	result, err := e.executor.Run(ctx, repo.Path, []string{"remote", "add", name, url})
	if err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to add remote %s: %s", name, result.Stderr)
	}

	e.logger.Info("Remote added", map[string]interface{}{
		"name": name,
		"url":  sanitizeURL(url),
	})

	return nil
}

func (e *ExecGit) RemoveRemote(ctx context.Context, repo *core.Repo, name string) error {
	if name == "" {
		return fmt.Errorf("remote name is required")
	}

	result, err := e.executor.Run(ctx, repo.Path, []string{"remote", "remove", name})
	if err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to remove remote %s: %s", name, result.Stderr)
	}

	e.logger.Info("Remote removed", map[string]interface{}{
		"name": name,
	})

	return nil
}

func (e *ExecGit) SetRemoteURL(ctx context.Context, repo *core.Repo, name, url string) error {
	if name == "" {
		return fmt.Errorf("remote name is required")
	}
	if url == "" {
		return fmt.Errorf("remote URL is required")
	}

	result, err := e.executor.Run(ctx, repo.Path, []string{"remote", "set-url", name, url})
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to set remote URL for %s: %s", name, result.Stderr)
	}

	e.logger.Info("Remote URL updated", map[string]interface{}{
		"name": name,
		"url":  sanitizeURL(url),
	})

	return nil
}

func (e *ExecGit) Fetch(ctx context.Context, repo *core.Repo, remote string, prune, tags bool) error {
	args := []string{"fetch"}
	
	if remote != "" {
		args = append(args, remote)
	}
	if prune {
		args = append(args, "--prune")
	}
	if tags {
		args = append(args, "--tags")
	}

	e.logger.Info("Fetching from remote", map[string]interface{}{
		"remote": remote,
		"prune":  prune,
		"tags":   tags,
	})

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("fetch failed: %s", result.Stderr)
	}

	return nil
}

func (e *ExecGit) Pull(ctx context.Context, repo *core.Repo, remote, branch string, rebase bool) error {
	args := []string{"pull"}
	
	if rebase {
		args = append(args, "--rebase")
	}
	if remote != "" {
		args = append(args, remote)
		if branch != "" {
			args = append(args, branch)
		}
	}

	e.logger.Info("Pulling from remote", map[string]interface{}{
		"remote": remote,
		"branch": branch,
		"rebase": rebase,
	})

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	if result.ExitCode != 0 {
		// Check for common error patterns
		if strings.Contains(result.Stderr, "merge conflict") {
			return fmt.Errorf("pull failed due to merge conflicts: resolve conflicts and commit")
		}
		if strings.Contains(result.Stderr, "non-fast-forward") {
			return fmt.Errorf("pull failed: non-fast-forward update rejected. Try pull --rebase or merge manually")
		}
		return fmt.Errorf("pull failed: %s", result.Stderr)
	}

	return nil
}

func (e *ExecGit) Push(ctx context.Context, repo *core.Repo, remote, branch string, force, tags bool) error {
	args := []string{"push"}
	
	if force {
		args = append(args, "--force-with-lease")
	}
	if tags {
		args = append(args, "--tags")
	}
	if remote != "" {
		args = append(args, remote)
		if branch != "" {
			args = append(args, branch)
		}
	}

	e.logger.Info("Pushing to remote", map[string]interface{}{
		"remote": remote,
		"branch": branch,
		"force":  force,
		"tags":   tags,
	})

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	if result.ExitCode != 0 {
		// Check for common error patterns
		if strings.Contains(result.Stderr, "non-fast-forward") {
			return fmt.Errorf("push rejected: non-fast-forward update. Use --force-with-lease if you're sure")
		}
		if strings.Contains(result.Stderr, "authentication") || strings.Contains(result.Stderr, "Permission denied") {
			return fmt.Errorf("push failed: authentication required. Check your credentials")
		}
		return fmt.Errorf("push failed: %s", result.Stderr)
	}

	return nil
}

func (e *ExecGit) CreateBranch(ctx context.Context, repo *core.Repo, name, startPoint string) error {
	if name == "" {
		return fmt.Errorf("branch name is required")
	}

	args := []string{"branch", name}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "already exists") {
			return fmt.Errorf("branch %s already exists", name)
		}
		return fmt.Errorf("failed to create branch %s: %s", name, result.Stderr)
	}

	e.logger.Info("Branch created", map[string]interface{}{
		"name":       name,
		"startPoint": startPoint,
	})

	return nil
}

func (e *ExecGit) DeleteBranch(ctx context.Context, repo *core.Repo, name string, force bool) error {
	if name == "" {
		return fmt.Errorf("branch name is required")
	}

	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, name)

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "not fully merged") {
			return fmt.Errorf("branch %s is not fully merged. Use force=true to delete anyway", name)
		}
		return fmt.Errorf("failed to delete branch %s: %s", name, result.Stderr)
	}

	e.logger.Info("Branch deleted", map[string]interface{}{
		"name":  name,
		"force": force,
	})

	return nil
}

func (e *ExecGit) Checkout(ctx context.Context, repo *core.Repo, ref string, createBranch bool) error {
	if ref == "" {
		return fmt.Errorf("reference is required")
	}

	args := []string{"checkout"}
	if createBranch {
		args = append(args, "-b")
	}
	args = append(args, ref)

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return fmt.Errorf("failed to checkout: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "already exists") {
			return fmt.Errorf("branch %s already exists", ref)
		}
		if strings.Contains(result.Stderr, "would be overwritten") {
			return fmt.Errorf("checkout failed: local changes would be overwritten. Commit or stash changes first")
		}
		return fmt.Errorf("checkout failed: %s", result.Stderr)
	}

	e.logger.Info("Checked out", map[string]interface{}{
		"ref":          ref,
		"createBranch": createBranch,
	})

	return nil
}

func (e *ExecGit) ListBranches(ctx context.Context, repo *core.Repo, all bool) ([]core.BranchInfo, error) {
	args := []string{"branch", "-v"}
	if all {
		args = append(args, "-a")
	}

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to list branches: %s", result.Stderr)
	}

	var branches []core.BranchInfo
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}

		line = strings.TrimSpace(line)
		current := strings.HasPrefix(line, "*")
		if current {
			line = strings.TrimSpace(line[1:])
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		// Skip remote tracking info in branch name
		if strings.Contains(name, "->") {
			continue
		}

		branch := core.BranchInfo{
			Name:    name,
			Current: current,
		}

		// Extract remote info for remote branches
		if strings.HasPrefix(name, "remotes/") {
			remoteParts := strings.SplitN(name[8:], "/", 2)
			if len(remoteParts) == 2 {
				branch.Remote = remoteParts[0]
				branch.Name = remoteParts[1]
			}
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func (e *ExecGit) Tag(ctx context.Context, repo *core.Repo, name, ref, message string, sign bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) DeleteTag(ctx context.Context, repo *core.Repo, name string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Merge(ctx context.Context, repo *core.Repo, ref string, noFF bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Rebase(ctx context.Context, repo *core.Repo, upstream string, interactive bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) CherryPick(ctx context.Context, repo *core.Repo, commit string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Revert(ctx context.Context, repo *core.Repo, commit string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Log(ctx context.Context, repo *core.Repo, ref string, maxCount int, oneline bool) ([]core.CommitInfo, error) {
	args := []string{"log"}
	
	if oneline {
		args = append(args, "--oneline")
	} else {
		args = append(args, "--pretty=format:%H|%h|%an|%ae|%ai|%s|%b")
	}
	
	if maxCount > 0 {
		args = append(args, "-n", strconv.Itoa(maxCount))
	}
	
	if ref != "" {
		args = append(args, ref)
	}

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("log failed: %s", result.Stderr)
	}

	var commits []core.CommitInfo
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}

		if oneline {
			// Parse oneline format: "hash subject"
			parts := strings.SplitN(line, " ", 2)
			if len(parts) >= 2 {
				commits = append(commits, core.CommitInfo{
					ShortHash: parts[0],
					Subject:   parts[1],
				})
			}
		} else {
			// Parse custom format: "hash|shorthash|author|email|date|subject|body"
			parts := strings.Split(line, "|")
			if len(parts) >= 6 {
				date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])
				body := ""
				if len(parts) > 6 {
					body = parts[6]
				}
				commits = append(commits, core.CommitInfo{
					Hash:      parts[0],
					ShortHash: parts[1],
					Author:    parts[2],
					Email:     parts[3],
					Date:      date,
					Subject:   parts[5],
					Body:      body,
				})
			}
		}
	}

	return commits, nil
}

func (e *ExecGit) Diff(ctx context.Context, repo *core.Repo, base, head string, stat bool) (string, error) {
	args := []string{"diff"}
	
	if stat {
		args = append(args, "--stat")
	}
	
	if base != "" && head != "" {
		args = append(args, base+"..."+head)
	} else if base != "" {
		args = append(args, base)
	} else if head != "" {
		args = append(args, head)
	}

	result, err := e.executor.Run(ctx, repo.Path, args)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("diff failed: %s", result.Stderr)
	}

	return result.Stdout, nil
}

func (e *ExecGit) Blame(ctx context.Context, repo *core.Repo, file, ref string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) RevParse(ctx context.Context, repo *core.Repo, ref string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Show(ctx context.Context, repo *core.Repo, ref string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) LsTree(ctx context.Context, repo *core.Repo, ref, path string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) StashSave(ctx context.Context, repo *core.Repo, message string, includeUntracked bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) StashList(ctx context.Context, repo *core.Repo) ([]string, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) StashPop(ctx context.Context, repo *core.Repo, index int) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) WorktreeCreate(ctx context.Context, repo *core.Repo, path, branch string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) WorktreeRemove(ctx context.Context, repo *core.Repo, path string, force bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) WorktreeList(ctx context.Context, repo *core.Repo) ([]string, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) SubmoduleInit(ctx context.Context, repo *core.Repo, path string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) SubmoduleUpdate(ctx context.Context, repo *core.Repo, path string, recursive bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) SubmoduleStatus(ctx context.Context, repo *core.Repo) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) LFSInstall(ctx context.Context, repo *core.Repo) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) LFSFetch(ctx context.Context, repo *core.Repo, remote string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) LFSPull(ctx context.Context, repo *core.Repo, remote string) error {
	return fmt.Errorf("not implemented yet")
}