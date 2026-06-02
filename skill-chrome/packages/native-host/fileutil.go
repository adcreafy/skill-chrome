package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func writeSkillFiles(skillsDir string, skillName string, files []resolvedFile) error {
	targetDir := filepath.Join(skillsDir, skillName)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	for _, f := range files {
		fullPath := filepath.Join(targetDir, f.RelativePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(f.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

type skillLockFile struct {
	Version  int                         `json:"version"`
	Skills   map[string]skillLockEntry   `json:"skills"`
	Dismissed map[string]json.RawMessage `json:"dismissed"`
}

type skillLockEntry struct {
	Source      string `json:"source"`
	SourceType  string `json:"sourceType"`
	SourceURL   string `json:"sourceUrl"`
	SkillPath   string `json:"skillPath"`
	InstalledAt string `json:"installedAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func updateLockFile(home, skillName, source, sourceType string) error {
	lockPath := filepath.Join(home, ".agents", ".skill-lock.json")

	var lock skillLockFile
	data, err := os.ReadFile(lockPath)
	if err == nil {
		_ = json.Unmarshal(data, &lock)
	}

	if lock.Version == 0 {
		lock.Version = 1
	}
	if lock.Skills == nil {
		lock.Skills = make(map[string]skillLockEntry)
	}
	if lock.Dismissed == nil {
		lock.Dismissed = make(map[string]json.RawMessage)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	existing, exists := lock.Skills[skillName]
	installedAt := now
	if exists {
		installedAt = existing.InstalledAt
	}

	lock.Skills[skillName] = skillLockEntry{
		Source:      source,
		SourceType:  sourceType,
		SourceURL:   source,
		SkillPath:   "skills/" + skillName + "/SKILL.md",
		InstalledAt: installedAt,
		UpdatedAt:   now,
	}

	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(lockPath, out, 0o644)
}
