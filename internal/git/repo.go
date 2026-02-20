package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository
type Repository struct {
	Path   string
	Owner  string
	Name   string
	Remote string
}

// DetectRepository detects the current Git repository
func DetectRepository(path string) (*Repository, error) {
	// Find .git directory
	gitDir, err := findGitDir(path)
	if err != nil {
		return nil, err
	}

	// Get remote URL
	cmd := exec.Command("git", "-C", gitDir, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, name, err := parseRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Path:   gitDir,
		Owner:  owner,
		Name:   name,
		Remote: remoteURL,
	}, nil
}

// GetCurrentRepository detects the repository for the current working directory
func GetCurrentRepository() (*Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return DetectRepository(cwd)
}

// Scope returns the scope identifier for this repository (e.g., "github.com/owner/repo")
func (r *Repository) Scope() string {
	return fmt.Sprintf("github.com/%s/%s", r.Owner, r.Name)
}

// findGitDir finds the .git directory starting from the given path
func findGitDir(startPath string) (string, error) {
	path := startPath
	for {
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return path, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("not a git repository")
		}
		path = parent
	}
}

// parseRemoteURL parses a Git remote URL to extract owner and repo name
// Supports both HTTPS and SSH formats:
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
func parseRemoteURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)
	
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL format: %s", url)
		}
		path := parts[1]
		pathParts := strings.Split(path, "/")
		if len(pathParts) != 2 {
			return "", "", fmt.Errorf("invalid repository path: %s", path)
		}
		return pathParts[0], pathParts[1], nil
	}

	// Handle HTTPS format: https://github.com/owner/repo
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		// Remove protocol
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		
		// Split by /
		parts := strings.Split(url, "/")
		if len(parts) < 3 {
			return "", "", fmt.Errorf("invalid HTTPS URL format: %s", url)
		}
		
		// parts[0] is domain (github.com), parts[1] is owner, parts[2] is repo
		return parts[1], parts[2], nil
	}

	return "", "", fmt.Errorf("unsupported URL format: %s", url)
}
