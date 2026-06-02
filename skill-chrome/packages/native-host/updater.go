package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	currentVersion = "0.1.0"
	repoOwner      = "adcreafy"
	repoName       = "skill-chrome"
	updateCheckInterval = 4 * time.Hour
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type updateState struct {
	LastCheck      string `json:"lastCheck"`
	LatestVersion  string `json:"latestVersion"`
	CurrentVersion string `json:"currentVersion"`
}

func stateFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skill-chrome", "update-state.json")
}

func shouldCheckUpdate() bool {
	data, err := os.ReadFile(stateFilePath())
	if err != nil {
		return true
	}
	var state updateState
	if err := json.Unmarshal(data, &state); err != nil {
		return true
	}
	lastCheck, err := time.Parse(time.RFC3339, state.LastCheck)
	if err != nil {
		return true
	}
	return time.Since(lastCheck) > updateCheckInterval
}

func saveUpdateState(latestVersion string) {
	state := updateState{
		LastCheck:      time.Now().UTC().Format(time.RFC3339),
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.MkdirAll(filepath.Dir(stateFilePath()), 0o755)
	_ = os.WriteFile(stateFilePath(), data, 0o644)
}

func checkForUpdate() (*githubRelease, bool) {
	if !shouldCheckUpdate() {
		return nil, false
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		saveUpdateState(currentVersion)
		return nil, false
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		saveUpdateState(currentVersion)
		return nil, false
	}

	latestTag := strings.TrimPrefix(release.TagName, "v")
	saveUpdateState(latestTag)

	if latestTag == currentVersion {
		return nil, false
	}

	if !isNewer(latestTag, currentVersion) {
		return nil, false
	}

	return &release, true
}

func isNewer(latest, current string) bool {
	lParts := strings.Split(latest, ".")
	cParts := strings.Split(current, ".")

	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		if lParts[i] > cParts[i] {
			return true
		}
		if lParts[i] < cParts[i] {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

func selfUpdate(release *githubRelease) error {
	suffix := runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, suffix) && strings.HasPrefix(asset.Name, "skill-chrome-host") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s", suffix)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Write to temp file first
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	tmpPath := execPath + ".new"
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmpPath)
		return err
	}
	out.Close()

	// Atomic replace: rename old, move new in place
	oldPath := execPath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Rollback
		_ = os.Rename(oldPath, execPath)
		return err
	}
	_ = os.Remove(oldPath)

	return nil
}

// tryBackgroundUpdate checks for updates and self-updates without blocking.
// Called from init() so it runs on every invocation, but rate-limited to
// check at most once per updateCheckInterval.
func tryBackgroundUpdate() {
	release, needsUpdate := checkForUpdate()
	if !needsUpdate || release == nil {
		return
	}
	// Perform update synchronously — native host is short-lived so this
	// only adds latency once when an update is actually available.
	_ = selfUpdate(release)
}
