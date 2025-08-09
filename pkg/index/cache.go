package index

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/felipemacedo1/go-coregit-pe/pkg/core"
)

// Cache provides lightweight caching for Git repository metadata
type Cache struct {
	basePath string
	mu       sync.RWMutex
}

// CacheEntry represents a cached item with TTL
type CacheEntry struct {
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".gitmgr", "cache")
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		basePath: basePath,
	}, nil
}

// getRepoHash generates a unique hash for a repository path
func (c *Cache) getRepoHash(repoPath string) string {
	hash := sha256.Sum256([]byte(repoPath))
	return fmt.Sprintf("%x", hash)
}

// getCacheDir returns the cache directory for a repository
func (c *Cache) getCacheDir(repoPath string) string {
	repoHash := c.getRepoHash(repoPath)
	return filepath.Join(c.basePath, repoHash)
}

// ensureCacheDir creates the cache directory for a repository if it doesn't exist
func (c *Cache) ensureCacheDir(repoPath string) error {
	cacheDir := c.getCacheDir(repoPath)
	return os.MkdirAll(cacheDir, 0755)
}

// Set stores data in cache with TTL
func (c *Cache) Set(repoPath, key string, data interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCacheDir(repoPath); err != nil {
		return err
	}

	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	cacheFile := filepath.Join(c.getCacheDir(repoPath), key+".json")
	if err := os.WriteFile(cacheFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Get retrieves data from cache
func (c *Cache) Get(repoPath, key string, target interface{}) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheFile := filepath.Join(c.getCacheDir(repoPath), key+".json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Cache miss
		}
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Entry expired, remove it
		_ = os.Remove(cacheFile)
		return false, nil
	}

	// Unmarshal the actual data
	entryData, err := json.Marshal(entry.Data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal entry data: %w", err)
	}

	if err := json.Unmarshal(entryData, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal target data: %w", err)
	}

	return true, nil
}

// Delete removes an entry from cache
func (c *Cache) Delete(repoPath, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheFile := filepath.Join(c.getCacheDir(repoPath), key+".json")
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// Clear removes all cache entries for a repository
func (c *Cache) Clear(repoPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheDir := c.getCacheDir(repoPath)
	if err := os.RemoveAll(cacheDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache directory: %w", err)
	}

	return nil
}

// CacheBranches stores branch information in cache
func (c *Cache) CacheBranches(repoPath string, branches []core.BranchInfo) error {
	return c.Set(repoPath, "branches", branches, 5*time.Minute)
}

// GetCachedBranches retrieves cached branch information
func (c *Cache) GetCachedBranches(repoPath string) ([]core.BranchInfo, bool, error) {
	var branches []core.BranchInfo
	found, err := c.Get(repoPath, "branches", &branches)
	return branches, found, err
}

// CacheRemotes stores remote information in cache
func (c *Cache) CacheRemotes(repoPath string, remotes []core.RemoteInfo) error {
	return c.Set(repoPath, "remotes", remotes, 10*time.Minute)
}

// GetCachedRemotes retrieves cached remote information
func (c *Cache) GetCachedRemotes(repoPath string) ([]core.RemoteInfo, bool, error) {
	var remotes []core.RemoteInfo
	found, err := c.Get(repoPath, "remotes", &remotes)
	return remotes, found, err
}

// CacheStatus stores repository status in cache
func (c *Cache) CacheStatus(repoPath string, status *core.RepoStatus) error {
	return c.Set(repoPath, "status", status, 30*time.Second)
}

// GetCachedStatus retrieves cached repository status
func (c *Cache) GetCachedStatus(repoPath string) (*core.RepoStatus, bool, error) {
	var status core.RepoStatus
	found, err := c.Get(repoPath, "status", &status)
	if err != nil {
		return nil, found, err
	}
	if !found {
		return nil, found, nil
	}
	return &status, found, err
}

// CacheCommits stores commit history in cache
func (c *Cache) CacheCommits(repoPath string, commits []core.CommitInfo) error {
	return c.Set(repoPath, "commits", commits, 2*time.Minute)
}

// GetCachedCommits retrieves cached commit history
func (c *Cache) GetCachedCommits(repoPath string) ([]core.CommitInfo, bool, error) {
	var commits []core.CommitInfo
	found, err := c.Get(repoPath, "commits", &commits)
	return commits, found, err
}
