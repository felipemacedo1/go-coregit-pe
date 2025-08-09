package executil

import (
	"context"
	"testing"
	"time"
)

func TestNewGitExecutor(t *testing.T) {
	executor := NewGitExecutor()
	if executor == nil {
		t.Fatal("NewGitExecutor returned nil")
	}
	if executor.timeout != 2*time.Minute {
		t.Errorf("Expected default timeout of 2 minutes, got %v", executor.timeout)
	}
}

func TestSetTimeout(t *testing.T) {
	executor := NewGitExecutor()
	newTimeout := 5 * time.Minute
	executor.SetTimeout(newTimeout)
	if executor.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, executor.timeout)
	}
}

func TestSanitizeArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "clean args",
			input:    []string{"status", "--porcelain"},
			expected: []string{"status", "--porcelain"},
		},
		{
			name:     "empty args",
			input:    []string{"", "status", ""},
			expected: []string{"status"},
		},
		{
			name:     "dangerous args",
			input:    []string{"status", "; rm -rf /"},
			expected: []string{"status"},
		},
		{
			name:     "safe message",
			input:    []string{"-m", "fix: update test"},
			expected: []string{"-m", "fix: update test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeArgs(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d", len(tt.expected), len(result))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("Expected arg %d to be %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean output",
			input:    "On branch main",
			expected: "On branch main",
		},
		{
			name:     "https credentials",
			input:    "fatal: Authentication failed for 'https://user:pass@github.com/repo.git'",
			expected: "fatal: Authentication failed for 'https://***:***@github.com/repo.git'",
		},
		{
			name:     "ssh user",
			input:    "ssh://git@github.com/repo.git",
			expected: "ssh://***@github.com/repo.git",
		},
		{
			name:     "token",
			input:    "token abc123def456",
			expected: "token ***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeOutput(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRun_InvalidRepo(t *testing.T) {
	executor := NewGitExecutor()
	ctx := context.Background()

	result, err := executor.Run(ctx, "/nonexistent/path", []string{"status"})
	if err != nil {
		// This is expected - git command will fail
		return
	}
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for nonexistent repository path")
	}
}
