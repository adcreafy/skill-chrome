package main

import "runtime"

// EngineDef defines a mainstream AI agent engine.
type EngineDef struct {
	ID         string
	Name       string
	DetectDirs []string // relative to home dir
	SkillsDir  string   // relative to home dir
}

// EngineFingerprint defines markers to identify apps built on a given engine.
type EngineFingerprint struct {
	EngineID        string
	ResourceMarkers []string // files/dirs expected under Contents/Resources/ (macOS) or install root (Windows)
}

// DetectedEngine is returned in the detect response.
type DetectedEngine struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Detected  bool          `json:"detected"`
	SkillsDir string        `json:"skillsPath"`
	Apps      []DetectedApp `json:"apps"`
}

// DetectedApp is a third-party app identified as wrapping a mainstream engine.
type DetectedApp struct {
	Name       string `json:"name"`
	BundlePath string `json:"bundlePath"`
}

func builtinEngines() []EngineDef {
	engines := []EngineDef{
		{ID: "cursor", Name: "Cursor", DetectDirs: []string{".cursor"}, SkillsDir: ".cursor/skills"},
		{ID: "claude-code", Name: "Claude Code", DetectDirs: []string{".claude"}, SkillsDir: ".claude/skills"},
		{ID: "openclaw", Name: "OpenClaw", DetectDirs: []string{".openclaw"}, SkillsDir: ".openclaw/skills"},
	}

	if runtime.GOOS == "windows" {
		engines = append(engines, EngineDef{
			ID: "hermes", Name: "Hermes",
			DetectDirs: []string{"AppData/Local/hermes"},
			SkillsDir:  "AppData/Local/hermes/skills",
		})
	} else {
		engines = append(engines, EngineDef{
			ID: "hermes", Name: "Hermes",
			DetectDirs: []string{".hermes"},
			SkillsDir:  ".hermes/skills",
		})
	}

	return engines
}

func builtinFingerprints() []EngineFingerprint {
	return []EngineFingerprint{
		{EngineID: "openclaw", ResourceMarkers: []string{"gateway.asar"}},
		{EngineID: "openclaw", ResourceMarkers: []string{"openclaw"}},
		{EngineID: "hermes", ResourceMarkers: []string{"hermes"}},
	}
}

// knownMainApps returns app names that ARE the engine itself (skip in scan).
func knownMainApps() map[string]bool {
	return map[string]bool{
		"cursor":   true,
		"openclaw": true,
		"hermes":   true,
		"claude":   true,
	}
}
