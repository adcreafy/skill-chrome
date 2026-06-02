package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSkillFiles(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")

	files := []resolvedFile{
		{RelativePath: "SKILL.md", Content: "# Test Skill\nA test skill.", Size: 26},
		{RelativePath: "sub/file.txt", Content: "hello", Size: 5},
	}

	if err := writeSkillFiles(skillsDir, "test-skill", files); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(skillsDir, "test-skill", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# Test Skill\nA test skill." {
		t.Errorf("unexpected content: %s", content)
	}

	subContent, err := os.ReadFile(filepath.Join(skillsDir, "test-skill", "sub", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(subContent) != "hello" {
		t.Errorf("unexpected sub content: %s", subContent)
	}
}

func TestUpdateLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := updateLockFile(tmpDir, "my-skill", "owner/repo", "github")
	if err != nil {
		t.Fatal(err)
	}

	lockPath := filepath.Join(tmpDir, ".agents", ".skill-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}

	var lock skillLockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatal(err)
	}

	if lock.Version != 1 {
		t.Errorf("expected version=1, got %d", lock.Version)
	}

	entry, ok := lock.Skills["my-skill"]
	if !ok {
		t.Fatal("expected my-skill in skills map")
	}
	if entry.Source != "owner/repo" {
		t.Errorf("expected source=owner/repo, got %s", entry.Source)
	}
	if entry.SourceType != "github" {
		t.Errorf("expected sourceType=github, got %s", entry.SourceType)
	}

	// Run again to test update (not duplicate)
	err = updateLockFile(tmpDir, "my-skill", "owner/repo", "github")
	if err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(lockPath)
	var lock2 skillLockFile
	json.Unmarshal(data2, &lock2)
	if len(lock2.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(lock2.Skills))
	}
}
