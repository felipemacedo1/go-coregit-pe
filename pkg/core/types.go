package core

import (
	"context"
	"time"
)

// Repo represents a Git repository
type Repo struct {
	Path       string
	WorkDir    string
	GitDir     string
	IsBare     bool
	IsWorktree bool
}

// CloneOptions configures repository cloning
type CloneOptions struct {
	URL       string
	Path      string
	Branch    string
	Depth     int
	Bare      bool
	Mirror    bool
	Sparse    []string
	Recursive bool
	Progress  bool
}

// ExecResult contains the result of a Git command execution
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// AuthHint provides authentication guidance
type AuthHint struct {
	Type    string // "ssh", "https", "token"
	Message string
	Helper  string
}

// BranchInfo represents branch information
type BranchInfo struct {
	Name     string
	Current  bool
	Remote   string
	Upstream string
	Ahead    int
	Behind   int
}

// RemoteInfo represents remote repository information
type RemoteInfo struct {
	Name     string
	URL      string
	FetchURL string
	PushURL  string
}

// CommitInfo represents commit information
type CommitInfo struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
}

// FileStatus represents file change status
type FileStatus struct {
	Path     string
	Status   string // "M", "A", "D", "R", "C", "U", "?", "!"
	Staged   bool
	Modified bool
}

// RepoStatus represents repository status
type RepoStatus struct {
	Branch   string
	Upstream string
	Ahead    int
	Behind   int
	Files    []FileStatus
	Clean    bool
}

// CoreGit defines the main interface for Git operations
type CoreGit interface {
	// Repository operations
	Open(ctx context.Context, path string) (*Repo, error)
	Init(ctx context.Context, path string, bare bool) (*Repo, error)
	Clone(ctx context.Context, opts CloneOptions) (*Repo, error)
	Discover(ctx context.Context, path string) (*Repo, error)
	GetStatus(ctx context.Context, repo *Repo) (*RepoStatus, error)
	GetConfig(ctx context.Context, repo *Repo, key string) (string, error)
	SetConfig(ctx context.Context, repo *Repo, key, value string, global bool) error

	// Remote operations
	ListRemotes(ctx context.Context, repo *Repo) ([]RemoteInfo, error)
	AddRemote(ctx context.Context, repo *Repo, name, url string) error
	RemoveRemote(ctx context.Context, repo *Repo, name string) error
	SetRemoteURL(ctx context.Context, repo *Repo, name, url string) error

	// Sync operations
	Fetch(ctx context.Context, repo *Repo, remote string, prune, tags bool) error
	Pull(ctx context.Context, repo *Repo, remote, branch string, rebase bool) error
	Push(ctx context.Context, repo *Repo, remote, branch string, force, tags bool) error

	// Branch and tag operations
	CreateBranch(ctx context.Context, repo *Repo, name, startPoint string) error
	DeleteBranch(ctx context.Context, repo *Repo, name string, force bool) error
	Checkout(ctx context.Context, repo *Repo, ref string, createBranch bool) error
	ListBranches(ctx context.Context, repo *Repo, all bool) ([]BranchInfo, error)
	Tag(ctx context.Context, repo *Repo, name, ref, message string, sign bool) error
	DeleteTag(ctx context.Context, repo *Repo, name string) error

	// Merge flow operations
	Merge(ctx context.Context, repo *Repo, ref string, noFF bool) error
	Rebase(ctx context.Context, repo *Repo, upstream string, interactive bool) error
	CherryPick(ctx context.Context, repo *Repo, commit string) error
	Revert(ctx context.Context, repo *Repo, commit string) error

	// Inspection operations
	Log(ctx context.Context, repo *Repo, ref string, maxCount int, oneline bool) ([]CommitInfo, error)
	Diff(ctx context.Context, repo *Repo, base, head string, stat bool) (string, error)
	Blame(ctx context.Context, repo *Repo, file, ref string) (string, error)
	RevParse(ctx context.Context, repo *Repo, ref string) (string, error)
	Show(ctx context.Context, repo *Repo, ref string) (string, error)
	LsTree(ctx context.Context, repo *Repo, ref, path string) (string, error)

	// Stash operations
	StashSave(ctx context.Context, repo *Repo, message string, includeUntracked bool) error
	StashList(ctx context.Context, repo *Repo) ([]string, error)
	StashPop(ctx context.Context, repo *Repo, index int) error

	// Worktree operations
	WorktreeCreate(ctx context.Context, repo *Repo, path, branch string) error
	WorktreeRemove(ctx context.Context, repo *Repo, path string, force bool) error
	WorktreeList(ctx context.Context, repo *Repo) ([]string, error)

	// Submodule operations
	SubmoduleInit(ctx context.Context, repo *Repo, path string) error
	SubmoduleUpdate(ctx context.Context, repo *Repo, path string, recursive bool) error
	SubmoduleStatus(ctx context.Context, repo *Repo) (string, error)

	// LFS operations
	LFSInstall(ctx context.Context, repo *Repo) error
	LFSFetch(ctx context.Context, repo *Repo, remote string) error
	LFSPull(ctx context.Context, repo *Repo, remote string) error

	// Raw command execution
	RunRaw(ctx context.Context, repo *Repo, args []string) (*ExecResult, error)
}
