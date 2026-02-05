package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
)

const cacheFileName = "tag_project_cache.json"

// Cache stores tags and projects for autocomplete
type Cache struct {
	Tags     []string `json:"tags"`
	Projects []string `json:"projects"`
}

// GetCachePath returns the path to the cache file
func GetCachePath() string {
	return filepath.Join(xdg.DataHome, "spendgrid", cacheFileName)
}

// LoadCache loads the cache from disk
func LoadCache() (*Cache, error) {
	cachePath := GetCachePath()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return &Cache{
			Tags:     []string{},
			Projects: []string{},
		}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %v", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %v", err)
	}

	return &cache, nil
}

// SaveCache saves the cache to disk
func (c *Cache) SaveCache() error {
	cachePath := GetCachePath()

	// Ensure directory exists
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %v", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %v", err)
	}

	return nil
}

// AddTag adds a tag to the cache if it doesn't exist
func (c *Cache) AddTag(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}

	// Check if already exists
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}

	c.Tags = append(c.Tags, tag)
	sort.Strings(c.Tags)
}

// AddProject adds a project to the cache if it doesn't exist
func (c *Cache) AddProject(project string) {
	project = strings.TrimSpace(project)
	if project == "" {
		return
	}

	// Check if already exists
	for _, p := range c.Projects {
		if p == project {
			return
		}
	}

	c.Projects = append(c.Projects, project)
	sort.Strings(c.Projects)
}

// GetTags returns all cached tags
func (c *Cache) GetTags() []string {
	return c.Tags
}

// GetProjects returns all cached projects
func (c *Cache) GetProjects() []string {
	return c.Projects
}

// GetMatchingTags returns tags that match the given prefix
func (c *Cache) GetMatchingTags(prefix string) []string {
	prefix = strings.ToLower(prefix)
	var matches []string

	for _, tag := range c.Tags {
		if strings.HasPrefix(strings.ToLower(tag), prefix) {
			matches = append(matches, tag)
		}
	}

	return matches
}

// GetMatchingProjects returns projects that match the given prefix
func (c *Cache) GetMatchingProjects(prefix string) []string {
	prefix = strings.ToLower(prefix)
	var matches []string

	for _, proj := range c.Projects {
		if strings.HasPrefix(strings.ToLower(proj), prefix) {
			matches = append(matches, proj)
		}
	}

	return matches
}
