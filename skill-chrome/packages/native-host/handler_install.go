package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type InstallPayload struct {
	Sources  []string `json:"sources"`
	AgentIDs []string `json:"agentIds"`
}

type InstallSucceeded struct {
	SkillName string   `json:"skillName"`
	Engines   []string `json:"engines"`
	FileCount int      `json:"fileCount"`
}

type InstallFailed struct {
	Source string `json:"source"`
	Error  string `json:"error"`
}

type InstallResult struct {
	Succeeded []InstallSucceeded `json:"succeeded"`
	Failed    []InstallFailed    `json:"failed"`
}

func handleInstall(raw json.RawMessage) Response {
	var payload InstallPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Response{OK: false, Error: "invalid payload: " + err.Error()}
	}

	if len(payload.Sources) == 0 {
		return Response{OK: false, Error: "no sources provided"}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Response{OK: false, Error: "cannot determine home directory"}
	}

	engines := builtinEngines()
	agentSet := make(map[string]bool, len(payload.AgentIDs))
	for _, id := range payload.AgentIDs {
		agentSet[id] = true
	}

	type resolveResult struct {
		skill *resolvedSkill
		err   error
		idx   int
	}

	ch := make(chan resolveResult, len(payload.Sources))
	var wg sync.WaitGroup

	for i, source := range payload.Sources {
		wg.Add(1)
		go func(idx int, src string) {
			defer wg.Done()
			skill, err := resolveSkill(src)
			ch <- resolveResult{skill: skill, err: err, idx: idx}
		}(i, source)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	result := InstallResult{
		Succeeded: []InstallSucceeded{},
		Failed:    []InstallFailed{},
	}

	for rr := range ch {
		if rr.err != nil {
			result.Failed = append(result.Failed, InstallFailed{
				Source: payload.Sources[rr.idx],
				Error:  rr.err.Error(),
			})
			continue
		}

		skill := rr.skill

		// Write to canonical ~/.agents/skills/
		canonicalDir := filepath.Join(home, ".agents", "skills")
		_ = writeSkillFiles(canonicalDir, skill.Name, skill.Files)

		// Write to each selected engine
		var installedEngines []string
		for _, eng := range engines {
			if !agentSet[eng.ID] {
				continue
			}
			engDir := filepath.Join(home, eng.SkillsDir)
			if err := writeSkillFiles(engDir, skill.Name, skill.Files); err == nil {
				installedEngines = append(installedEngines, eng.ID)
			}
		}

		// Update lock file
		_ = updateLockFile(home, skill.Name, skill.Source, skill.SourceType)

		result.Succeeded = append(result.Succeeded, InstallSucceeded{
			SkillName: skill.Name,
			Engines:   installedEngines,
			FileCount: len(skill.Files),
		})
	}

	return Response{OK: true, Data: result}
}
