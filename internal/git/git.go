package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// IsRepo checks if a directory is a git repository
func IsRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// GetRemote gets the remote URL for a git repository
func GetRemote(path string) string {
	cmd := exec.Command("git", "-C", path, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetBranch gets the current git branch
func GetBranch(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetModifiedFiles gets a list of modified files in git
func GetModifiedFiles(path string) ([]string, error) {
	if !IsRepo(path) {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "-C", path, "diff", "--name-only", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Convert to absolute paths
	for i, file := range files {
		files[i] = filepath.Join(path, file)
	}

	return files, nil
}

// GetStagedFiles gets a list of staged files in git
func GetStagedFiles(path string) ([]string, error) {
	if !IsRepo(path) {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "-C", path, "diff", "--name-only", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Convert to absolute paths
	for i, file := range files {
		files[i] = filepath.Join(path, file)
	}

	return files, nil
}

// ParseGitignore parses a .gitignore file into a regexp pattern
func ParseGitignore(path string) (*regexp.Regexp, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var patterns []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Convert gitignore pattern to regex
		pattern := gitignoreToRegexp(line)
		patterns = append(patterns, pattern)
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	// Join all patterns with OR
	regexPattern := fmt.Sprintf("(%s)", strings.Join(patterns, "|"))
	return regexp.Compile(regexPattern)
}

// gitignoreToRegexp converts a gitignore pattern to a regular expression
func gitignoreToRegexp(pattern string) string {
	// Escape special regex characters
	pattern = regexp.QuoteMeta(pattern)

	// Convert gitignore glob patterns to regex
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")

	// Handle directory separator
	if strings.HasSuffix(pattern, "/") {
		pattern = pattern + ".*"
	}

	// Handle negation (!)
	if strings.HasPrefix(pattern, "\\!") {
		pattern = pattern[2:]
	}

	// Handle directory-only pattern (*/)
	pattern = strings.ReplaceAll(pattern, ".*/", ".*\\/")

	return pattern
}

// GetFileDiff gets the diff for a specific file
func GetFileDiff(repoPath, filePath string) (string, error) {
	if !IsRepo(repoPath) {
		return "", fmt.Errorf("not a git repository")
	}

	// Get relative path from repo root
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "-C", repoPath, "diff", "HEAD", "--", relPath)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// GetRepoSummary gets a summary of the git repository
func GetRepoSummary(path string) (map[string]string, error) {
	if !IsRepo(path) {
		return nil, fmt.Errorf("not a git repository")
	}

	summary := make(map[string]string)

	// Get remote URL
	remote := GetRemote(path)
	if remote != "" {
		summary["remote"] = remote
	}

	// Get current branch
	branch := GetBranch(path)
	if branch != "" {
		summary["branch"] = branch
	}

	// Get commit count
	cmd := exec.Command("git", "-C", path, "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		summary["commits"] = strings.TrimSpace(string(out))
	}

	// Get last commit info
	cmd = exec.Command("git", "-C", path, "log", "-1", "--pretty=format:%h %s (%cr)")
	out, err = cmd.Output()
	if err == nil {
		summary["last_commit"] = strings.TrimSpace(string(out))
	}

	// Get modified files count
	modifiedFiles, err := GetModifiedFiles(path)
	if err == nil {
		summary["modified_files"] = fmt.Sprintf("%d", len(modifiedFiles))
	}

	// Get staged files count
	stagedFiles, err := GetStagedFiles(path)
	if err == nil {
		summary["staged_files"] = fmt.Sprintf("%d", len(stagedFiles))
	}

	return summary, nil
}