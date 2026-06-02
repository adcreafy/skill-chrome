package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectResult is the data payload for a detect_agents response.
type DetectResult struct {
	Engines []DetectedEngine `json:"engines"`
}

func handleDetect() Response {
	home, err := os.UserHomeDir()
	if err != nil {
		return Response{OK: false, Error: "cannot determine home directory: " + err.Error()}
	}

	engines := builtinEngines()
	result := make([]DetectedEngine, 0, len(engines))

	for _, eng := range engines {
		detected := false
		for _, dir := range eng.DetectDirs {
			fullPath := filepath.Join(home, dir)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				detected = true
				break
			}
		}
		result = append(result, DetectedEngine{
			ID:        eng.ID,
			Name:      eng.Name,
			Detected:  detected,
			SkillsDir: filepath.Join(home, eng.SkillsDir),
			Apps:      []DetectedApp{},
		})
	}

	scannedApps := scanApplications()
	for i := range result {
		for _, app := range scannedApps {
			if app.engineID == result[i].ID {
				result[i].Apps = append(result[i].Apps, DetectedApp{
					Name:       app.name,
					BundlePath: app.bundlePath,
				})
			}
		}
	}

	return Response{OK: true, Data: DetectResult{Engines: result}}
}

type scannedApp struct {
	name       string
	bundlePath string
	engineID   string
}

func scanApplications() []scannedApp {
	if runtime.GOOS == "darwin" {
		return scanMacApps()
	}
	if runtime.GOOS == "windows" {
		return scanWindowsApps()
	}
	return nil
}

func scanMacApps() []scannedApp {
	appsDir := "/Applications"
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil
	}

	fingerprints := builtinFingerprints()
	mainApps := knownMainApps()
	var results []scannedApp

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
			continue
		}

		appPath := filepath.Join(appsDir, entry.Name())
		resourcesPath := filepath.Join(appPath, "Contents", "Resources")

		resourceEntries, err := os.ReadDir(resourcesPath)
		if err != nil {
			continue
		}

		resourceNames := make(map[string]bool)
		for _, re := range resourceEntries {
			resourceNames[strings.ToLower(re.Name())] = true
		}

		for _, fp := range fingerprints {
			allFound := true
			for _, marker := range fp.ResourceMarkers {
				if !resourceNames[strings.ToLower(marker)] {
					allFound = false
					break
				}
			}
			if !allFound {
				continue
			}

			appName := strings.TrimSuffix(entry.Name(), ".app")

			if mainApps[strings.ToLower(appName)] {
				continue
			}

			name := parsePlistDisplayName(appPath)
			if name == "" {
				name = appName
			}

			results = append(results, scannedApp{
				name:       name,
				bundlePath: appPath,
				engineID:   fp.EngineID,
			})
			break
		}
	}

	return results
}

func scanWindowsApps() []scannedApp {
	// Windows app scanning: check Program Files and LocalAppData
	home, _ := os.UserHomeDir()
	scanDirs := []string{
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		filepath.Join(home, "AppData", "Local", "Programs"),
	}

	fingerprints := builtinFingerprints()
	mainApps := knownMainApps()
	var results []scannedApp

	for _, scanDir := range scanDirs {
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			appPath := filepath.Join(scanDir, entry.Name())
			resourcesPath := filepath.Join(appPath, "resources")

			resourceEntries, err := os.ReadDir(resourcesPath)
			if err != nil {
				continue
			}

			resourceNames := make(map[string]bool)
			for _, re := range resourceEntries {
				resourceNames[strings.ToLower(re.Name())] = true
			}

			for _, fp := range fingerprints {
				allFound := true
				for _, marker := range fp.ResourceMarkers {
					if !resourceNames[strings.ToLower(marker)] {
						allFound = false
						break
					}
				}
				if !allFound {
					continue
				}
				if mainApps[strings.ToLower(entry.Name())] {
					continue
				}
				results = append(results, scannedApp{
					name:       entry.Name(),
					bundlePath: appPath,
					engineID:   fp.EngineID,
				})
				break
			}
		}
	}

	return results
}
