package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func IsRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func GetRemote(path string) string {
	cmd := exec.Command("git", "-C", path, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetBranch(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ParseGitignore(path string) (*regexp.Regexp, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var patterns []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pat := regexp.QuoteMeta(line)
		pat = strings.ReplaceAll(pat, "\\*", ".*")
		patterns = append(patterns, pat)
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return regexp.Compile(strings.Join(patterns, "|"))
}
