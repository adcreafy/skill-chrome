package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractNameHint(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"owner/my-skill", "my-skill"},
		{"https://github.com/owner/my-skill", "my-skill"},
		{"https://github.com/owner/my-skill/", "my-skill"},
		{"simple-name", "simple-name"},
	}
	for _, c := range cases {
		got := extractNameHint(c.input)
		if got != c.expected {
			t.Errorf("extractNameHint(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestPickBestMatch(t *testing.T) {
	dirs := []string{"/tmp/skills/foo", "/tmp/skills/bar-skill", "/tmp/skills/baz"}

	got := pickBestMatch(dirs, "bar-skill")
	if got != "/tmp/skills/bar-skill" {
		t.Errorf("expected bar-skill dir, got %s", got)
	}

	got2 := pickBestMatch(dirs, "bar")
	if got2 != "/tmp/skills/bar-skill" {
		t.Errorf("expected bar-skill dir via partial, got %s", got2)
	}

	got3 := pickBestMatch(dirs, "nonexistent")
	if got3 != "/tmp/skills/baz" {
		t.Errorf("expected last dir fallback, got %s", got3)
	}
}

func TestInferSourceType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"owner/repo", "github"},
		{"https://github.com/owner/repo", "github"},
		{"https://skills.sh/my-skill", "registry"},
		{"https://example.com/skill", "url"},
	}
	for _, c := range cases {
		got := inferSourceType(c.input)
		if got != c.expected {
			t.Errorf("inferSourceType(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestScanDir(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte("# Skill"), 0o644)
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "sub", "util.js"), []byte("export default {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte("{}"), 0o644)

	files, err := scanDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	names := make(map[string]bool)
	for _, f := range files {
		names[f.RelativePath] = true
	}

	if !names["SKILL.md"] {
		t.Error("expected SKILL.md")
	}
	if !names["sub/util.js"] {
		t.Error("expected sub/util.js")
	}
	if names["metadata.json"] {
		t.Error("metadata.json should be excluded")
	}
}

func TestParseFrontmatterDesc(t *testing.T) {
	files := []resolvedFile{
		{
			RelativePath: "SKILL.md",
			Content: `---
name: test-skill
description: "A test description"
---
# Test Skill`,
		},
	}
	desc := parseFrontmatterDesc(files)
	if desc != "A test description" {
		t.Errorf("expected 'A test description', got %q", desc)
	}
}
