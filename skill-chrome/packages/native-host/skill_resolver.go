package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type resolvedFile struct {
	RelativePath string `json:"relativePath"`
	Content      string `json:"content"`
	Size         int    `json:"size"`
}

type resolvedSkill struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Source      string         `json:"source"`
	SourceType  string         `json:"sourceType"`
	Files       []resolvedFile `json:"files"`
}

func resolveSkill(source string) (*resolvedSkill, error) {
	tmpDir, err := os.MkdirTemp("", "skill-installer-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Try npx first if available
	if npxPath, err := exec.LookPath("npx"); err == nil {
		if skill, err := resolveViaNpx(npxPath, source, tmpDir); err == nil {
			return skill, nil
		}
	}

	// Fallback to git clone
	if gitPath, err := exec.LookPath("git"); err == nil {
		if skill, err := resolveViaGit(gitPath, source, tmpDir); err == nil {
			return skill, nil
		}
	}

	return nil, fmt.Errorf("cannot resolve skill %q: npx and git not available or failed", source)
}

func resolveViaNpx(npxPath, source, tmpDir string) (*resolvedSkill, error) {
	cmd := exec.Command(npxPath, "skills", "add", source, "--copy", "-y")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return pickSkillFromTmp(tmpDir, source)
}

// githubTreeRe matches URLs like https://github.com/owner/repo/tree/branch/path/to/dir
var githubTreeRe = regexp.MustCompile(`^https?://github\.com/([^/]+/[^/]+)/tree/([^/]+)(?:/(.+))?$`)

// parseGitHubSource extracts repo URL, branch, and subpath from a GitHub URL.
// Returns (repoURL, branch, subpath). For non-tree URLs, branch and subpath are empty.
func parseGitHubSource(source string) (repoURL, branch, subpath string) {
	m := githubTreeRe.FindStringSubmatch(source)
	if m != nil {
		return "https://github.com/" + m[1] + ".git", m[2], m[3]
	}

	// Plain GitHub URL like https://github.com/owner/repo
	if strings.HasPrefix(source, "https://github.com/") || strings.HasPrefix(source, "http://github.com/") {
		return source + ".git", "", ""
	}

	// Short form: owner/repo
	if !strings.HasPrefix(source, "http") {
		return "https://github.com/" + source + ".git", "", ""
	}

	return source, "", ""
}

func resolveViaGit(gitPath, source, tmpDir string) (*resolvedSkill, error) {
	repoURL, branch, subpath := parseGitHubSource(source)

	cloneDir := filepath.Join(tmpDir, "repo")

	args := []string{"clone", "--depth", "1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, cloneDir)

	cmd := exec.Command(gitPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone failed: %s: %s", err, string(out))
	}

	// If subpath specified, look for SKILL.md starting from that subpath
	searchRoot := cloneDir
	if subpath != "" {
		searchRoot = filepath.Join(cloneDir, subpath)
		if _, err := os.Stat(searchRoot); err != nil {
			return nil, fmt.Errorf("subpath %q not found in repo", subpath)
		}
	}

	// Check if the search root itself contains SKILL.md
	if _, err := os.Stat(filepath.Join(searchRoot, "SKILL.md")); err == nil {
		files, err := scanDir(searchRoot)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(searchRoot)
		desc := parseFrontmatterDesc(files)
		return &resolvedSkill{
			Name:        name,
			Description: desc,
			Source:      source,
			SourceType:  inferSourceType(source),
			Files:       files,
		}, nil
	}

	// Walk to find SKILL.md
	var skillDir string
	_ = filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "SKILL.md" && !info.IsDir() {
			skillDir = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})

	if skillDir == "" {
		return nil, fmt.Errorf("no SKILL.md found in %s", source)
	}

	files, err := scanDir(skillDir)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(skillDir)
	desc := parseFrontmatterDesc(files)
	return &resolvedSkill{
		Name:        name,
		Description: desc,
		Source:      source,
		SourceType:  inferSourceType(source),
		Files:       files,
	}, nil
}

func pickSkillFromTmp(tmpDir, source string) (*resolvedSkill, error) {
	searchPaths := []string{
		filepath.Join(tmpDir, ".agents", "skills"),
		filepath.Join(tmpDir, ".cursor", "skills"),
		filepath.Join(tmpDir, ".claude", "skills"),
		filepath.Join(tmpDir, ".openclaw", "skills"),
	}

	hint := extractNameHint(source)
	var allDirs []string

	for _, sp := range searchPaths {
		entries, err := os.ReadDir(sp)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			skillMd := filepath.Join(sp, e.Name(), "SKILL.md")
			if _, err := os.Stat(skillMd); err == nil {
				allDirs = append(allDirs, filepath.Join(sp, e.Name()))
			}
		}
	}

	if len(allDirs) == 0 {
		return nil, fmt.Errorf("no skill directories found after npx skills add")
	}

	skillDir := pickBestMatch(allDirs, hint)
	files, err := scanDir(skillDir)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(skillDir)
	desc := parseFrontmatterDesc(files)
	return &resolvedSkill{
		Name:        name,
		Description: desc,
		Source:      source,
		SourceType:  inferSourceType(source),
		Files:       files,
	}, nil
}

func extractNameHint(source string) string {
	cleaned := strings.TrimRight(source, "/")
	parts := strings.Split(cleaned, "/")
	return parts[len(parts)-1]
}

func pickBestMatch(dirs []string, hint string) string {
	if hint == "" || len(dirs) == 1 {
		return dirs[len(dirs)-1]
	}
	lower := strings.ToLower(hint)
	for _, d := range dirs {
		if strings.ToLower(filepath.Base(d)) == lower {
			return d
		}
	}
	for _, d := range dirs {
		if strings.Contains(strings.ToLower(filepath.Base(d)), lower) {
			return d
		}
	}
	return dirs[len(dirs)-1]
}

func scanDir(dir string) ([]resolvedFile, error) {
	var files []resolvedFile
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := info.Name()
		if info.IsDir() && (base == "node_modules" || base == "__pycache__" || base == ".git") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if base == "metadata.json" || base == "skills-lock.json" || base == ".skill-lock.json" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		files = append(files, resolvedFile{
			RelativePath: filepath.ToSlash(rel),
			Content:      string(content),
			Size:         len(content),
		})
		return nil
	})
	return files, err
}

var frontmatterRe = regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
var descRe = regexp.MustCompile(`(?m)^description:\s*["']?(.+?)["']?\s*$`)

func parseFrontmatterDesc(files []resolvedFile) string {
	for _, f := range files {
		if f.RelativePath == "SKILL.md" {
			m := frontmatterRe.FindStringSubmatch(f.Content)
			if m == nil {
				return ""
			}
			dm := descRe.FindStringSubmatch(m[1])
			if dm != nil {
				return dm[1]
			}
			return ""
		}
	}
	return ""
}

func inferSourceType(source string) string {
	if strings.Contains(source, "github.com") || regexp.MustCompile(`^[\w-]+/[\w-]+$`).MatchString(source) {
		return "github"
	}
	if strings.Contains(source, "skills.sh") {
		return "registry"
	}
	return "url"
}
