package execgit

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

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
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Discover(ctx context.Context, path string) (*core.Repo, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) GetConfig(ctx context.Context, repo *core.Repo, key string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (e *ExecGit) SetConfig(ctx context.Context, repo *core.Repo, key, value string, global bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) ListRemotes(ctx context.Context, repo *core.Repo) ([]core.RemoteInfo, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) AddRemote(ctx context.Context, repo *core.Repo, name, url string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) RemoveRemote(ctx context.Context, repo *core.Repo, name string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) SetRemoteURL(ctx context.Context, repo *core.Repo, name, url string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Fetch(ctx context.Context, repo *core.Repo, remote string, prune, tags bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Pull(ctx context.Context, repo *core.Repo, remote, branch string, rebase bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Push(ctx context.Context, repo *core.Repo, remote, branch string, force, tags bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) CreateBranch(ctx context.Context, repo *core.Repo, name, startPoint string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) DeleteBranch(ctx context.Context, repo *core.Repo, name string, force bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Checkout(ctx context.Context, repo *core.Repo, ref string, createBranch bool) error {
	return fmt.Errorf("not implemented yet")
}

func (e *ExecGit) ListBranches(ctx context.Context, repo *core.Repo, all bool) ([]core.BranchInfo, error) {
	return nil, fmt.Errorf("not implemented yet")
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
	return nil, fmt.Errorf("not implemented yet")
}

func (e *ExecGit) Diff(ctx context.Context, repo *core.Repo, base, head string, stat bool) (string, error) {
	return "", fmt.Errorf("not implemented yet")
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