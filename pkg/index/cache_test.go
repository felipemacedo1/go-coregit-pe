package index

import (
	"testing"
	"time"

	"github.com/felipe-macedo/go-coregit-pe/pkg/core"
)

func TestNewCache(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
	if cache.basePath == "" {
		t.Error("basePath is empty")
	}
}

func TestCacheSetGet(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	repoPath := "/test/repo"
	key := "test-key"
	data := map[string]string{"test": "value"}

	// Set data
	err = cache.Set(repoPath, key, data, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get data
	var result map[string]string
	found, err := cache.Get(repoPath, key, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Error("Expected to find cached data")
	}
	if result["test"] != "value" {
		t.Errorf("Expected 'value', got '%s'", result["test"])
	}

	// Clean up
	cache.Clear(repoPath)
}

func TestCacheExpiration(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	repoPath := "/test/repo"
	key := "test-key"
	data := "test-data"

	// Set data with very short TTL
	err = cache.Set(repoPath, key, data, time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to get expired data
	var result string
	found, err := cache.Get(repoPath, key, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if found {
		t.Error("Expected expired data to not be found")
	}

	// Clean up
	cache.Clear(repoPath)
}

func TestCacheBranches(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	repoPath := "/test/repo"
	branches := []core.BranchInfo{
		{Name: "main", Current: true},
		{Name: "develop", Current: false},
	}

	// Cache branches
	err = cache.CacheBranches(repoPath, branches)
	if err != nil {
		t.Fatalf("CacheBranches failed: %v", err)
	}

	// Get cached branches
	result, found, err := cache.GetCachedBranches(repoPath)
	if err != nil {
		t.Fatalf("GetCachedBranches failed: %v", err)
	}
	if !found {
		t.Error("Expected to find cached branches")
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(result))
	}
	if result[0].Name != "main" || !result[0].Current {
		t.Error("First branch data incorrect")
	}

	// Clean up
	cache.Clear(repoPath)
}

func TestCacheDelete(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	repoPath := "/test/repo"
	key := "test-key"
	data := "test-data"

	// Set data
	err = cache.Set(repoPath, key, data, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete data
	err = cache.Delete(repoPath, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to get deleted data
	var result string
	found, err := cache.Get(repoPath, key, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if found {
		t.Error("Expected deleted data to not be found")
	}

	// Clean up
	cache.Clear(repoPath)
}

func TestGetRepoHash(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	path1 := "/test/repo1"
	path2 := "/test/repo2"

	hash1 := cache.getRepoHash(path1)
	hash2 := cache.getRepoHash(path2)

	if hash1 == hash2 {
		t.Error("Different paths should have different hashes")
	}
	if len(hash1) != 64 { // SHA256 hex length
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	// Same path should produce same hash
	hash1_again := cache.getRepoHash(path1)
	if hash1 != hash1_again {
		t.Error("Same path should produce same hash")
	}
}