package execgit

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	git := New()
	if git == nil {
		t.Fatal("New() returned nil")
	}
	if git.executor == nil {
		t.Error("executor is nil")
	}
	if git.logger == nil {
		t.Error("logger is nil")
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean https url",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "https with credentials",
			input:    "https://user:token@github.com/user/repo.git",
			expected: "https://***@github.com/user/repo.git",
		},
		{
			name:     "ssh url",
			input:    "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "http with credentials",
			input:    "http://user:pass@example.com/repo.git",
			expected: "http://***@example.com/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestOpen_InvalidPath(t *testing.T) {
	git := New()
	ctx := context.Background()

	_, err := git.Open(ctx, "/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}
