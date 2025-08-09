package executil

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// GitExecutor handles secure execution of git commands
type GitExecutor struct {
	timeout time.Duration
}

// NewGitExecutor creates a new GitExecutor with default timeout
func NewGitExecutor() *GitExecutor {
	return &GitExecutor{
		timeout: 2 * time.Minute,
	}
}

// SetTimeout configures the default timeout for git commands
func (e *GitExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// ExecResult contains the result of command execution
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Run executes a git command with security measures
func (e *GitExecutor) Run(ctx context.Context, repoPath string, args []string) (*ExecResult, error) {
	start := time.Now()
	
	// Create context with timeout if none provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
		defer cancel()
	}

	// Sanitize arguments to prevent injection
	sanitizedArgs := sanitizeArgs(args)
	
	// Build command
	cmdArgs := []string{"-C", repoPath}
	cmdArgs = append(cmdArgs, sanitizedArgs...)
	
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	
	// Set secure environment
	cmd.Env = []string{
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C",
		"PATH=" + getSecurePath(),
	}
	
	// Execute command
	stdout, err := cmd.Output()
	stderr := ""
	exitCode := 0
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = string(exitError.Stderr)
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute git command: %w", err)
		}
	}
	
	result := &ExecResult{
		ExitCode: exitCode,
		Stdout:   string(stdout),
		Stderr:   sanitizeOutput(stderr),
		Duration: time.Since(start),
	}
	
	return result, nil
}

// sanitizeArgs removes potentially dangerous arguments
func sanitizeArgs(args []string) []string {
	var sanitized []string
	for _, arg := range args {
		// Skip empty arguments
		if arg == "" {
			continue
		}
		// Basic validation - no shell metacharacters in unexpected places
		if strings.ContainsAny(arg, ";|&$`") && !isKnownSafeArg(arg) {
			continue
		}
		sanitized = append(sanitized, arg)
	}
	return sanitized
}

// isKnownSafeArg checks if an argument with special characters is known to be safe
func isKnownSafeArg(arg string) bool {
	// Allow commit messages and other text that might contain special chars
	safePatterns := []string{
		`^-m$`,           // message flag
		`^--message=`,    // message with value
		`^--format=`,     // format string
		`^--pretty=`,     // pretty format
	}
	
	for _, pattern := range safePatterns {
		if matched, _ := regexp.MatchString(pattern, arg); matched {
			return true
		}
	}
	return false
}

// sanitizeOutput removes sensitive information from command output
func sanitizeOutput(output string) string {
	// Regex patterns for common sensitive data
	patterns := []struct {
		pattern     string
		replacement string
	}{
		{`https://[^:]+:[^@]+@`, "https://***:***@"},           // HTTPS credentials
		{`ssh://[^@]+@`, "ssh://***@"},                         // SSH user
		{`token [a-zA-Z0-9_-]+`, "token ***"},                  // Generic tokens
		{`password[=:]\s*[^\s]+`, "password=***"},              // Password fields
	}
	
	result := output
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replacement)
	}
	
	return result
}

// getSecurePath returns a minimal PATH for git execution
func getSecurePath() string {
	// Common git installation paths
	paths := []string{
		"/usr/bin",
		"/usr/local/bin",
		"/opt/homebrew/bin", // macOS Homebrew
		"/bin",
	}
	return strings.Join(paths, ":")
}