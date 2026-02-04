package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RecentDir represents a recently used SpendGrid directory
type RecentDir struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	LastUsed time.Time `json:"last_used"`
}

// RecentDirsStore manages the list of recent directories
type RecentDirsStore struct {
	Directories []RecentDir `json:"directories"`
	maxSize     int
}

const maxRecentDirs = 10
const recentDirsFile = "recent_directories.json"

// GetRecentDirsStore loads or creates the recent directories store
func GetRecentDirsStore() (*RecentDirsStore, error) {
	store := &RecentDirsStore{
		Directories: []RecentDir{},
		maxSize:     maxRecentDirs,
	}

	filePath := filepath.Join(dataPath, recentDirsFile)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist yet, return empty store
		return store, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read recent directories: %v", err)
	}

	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("failed to parse recent directories: %v", err)
	}

	return store, nil
}

// AddDirectory adds a new directory to the recent list
func (s *RecentDirsStore) AddDirectory(dirPath string) error {
	// Get directory name (last part of path)
	dirName := filepath.Base(dirPath)

	// Check if already exists
	for i, dir := range s.Directories {
		if dir.Path == dirPath {
			// Update last used time and move to front
			s.Directories[i].LastUsed = time.Now()
			s.moveToFront(i)
			return s.Save()
		}
	}

	// Add new directory
	newDir := RecentDir{
		Path:     dirPath,
		Name:     dirName,
		LastUsed: time.Now(),
	}

	// Add to front
	s.Directories = append([]RecentDir{newDir}, s.Directories...)

	// Trim to max size
	if len(s.Directories) > s.maxSize {
		s.Directories = s.Directories[:s.maxSize]
	}

	return s.Save()
}

// GetDirectories returns the list of recent directories
func (s *RecentDirsStore) GetDirectories() []RecentDir {
	return s.Directories
}

// Save persists the store to disk
func (s *RecentDirsStore) Save() error {
	filePath := filepath.Join(dataPath, recentDirsFile)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal recent directories: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write recent directories: %v", err)
	}

	return nil
}

// moveToFront moves an item at index i to the front of the slice
func (s *RecentDirsStore) moveToFront(i int) {
	if i == 0 {
		return
	}
	item := s.Directories[i]
	copy(s.Directories[1:i+1], s.Directories[:i])
	s.Directories[0] = item
}

// SaveCurrentDirectory saves the current working directory to recent list
func SaveCurrentDirectory() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Only save if it's a SpendGrid directory
	if _, err := os.Stat(filepath.Join(cwd, ".spendgrid")); os.IsNotExist(err) {
		return nil // Not a SpendGrid directory, don't save
	}

	store, err := GetRecentDirsStore()
	if err != nil {
		return err
	}

	return store.AddDirectory(cwd)
}
