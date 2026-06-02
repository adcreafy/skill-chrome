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

func resolveViaGit(gitPath, source, tmpDir string) (*resolvedSkill, error) {
	repoURL := source
	if !strings.HasPrefix(source, "http") {
		repoURL = "https://github.com/" + source + ".git"
	}
	cloneDir := filepath.Join(tmpDir, "repo")
	cmd := exec.Command(gitPath, "clone", "--depth", "1", repoURL, cloneDir)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Find SKILL.md in cloned repo
	var skillDir string
	_ = filepath.Walk(cloneDir, func(path string, info os.FileInfo, err error) error {
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
