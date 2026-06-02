package main

import (
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	expandPath()
}

// expandPath ensures common tool directories are in PATH.
// Chrome launches native hosts with a very minimal PATH that often
// excludes Homebrew, nvm, and other user-installed tool locations.
func expandPath() {
	home, _ := os.UserHomeDir()
	if home == "" {
		return
	}

	extra := []string{}
	if runtime.GOOS == "darwin" {
		extra = append(extra,
			"/opt/homebrew/bin",
			"/usr/local/bin",
			home+"/.nvm/versions/node/current/bin",
		)
		// Detect actual nvm node versions
		nvmDir := home + "/.nvm/versions/node"
		if entries, err := os.ReadDir(nvmDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && e.Name() != "current" {
					extra = append(extra, nvmDir+"/"+e.Name()+"/bin")
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		extra = append(extra,
			"/usr/local/bin",
			"/snap/bin",
			home+"/.local/bin",
			home+"/.nvm/versions/node/current/bin",
		)
		nvmDir := home + "/.nvm/versions/node"
		if entries, err := os.ReadDir(nvmDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && e.Name() != "current" {
					extra = append(extra, nvmDir+"/"+e.Name()+"/bin")
				}
			}
		}
	}

	if len(extra) == 0 {
		return
	}

	current := os.Getenv("PATH")
	existing := make(map[string]bool)
	for _, p := range strings.Split(current, string(os.PathListSeparator)) {
		existing[p] = true
	}

	var toAdd []string
	for _, p := range extra {
		if !existing[p] {
			if _, err := os.Stat(p); err == nil {
				toAdd = append(toAdd, p)
			}
		}
	}

	if len(toAdd) > 0 {
		os.Setenv("PATH", current+string(os.PathListSeparator)+strings.Join(toAdd, string(os.PathListSeparator)))
	}
}

func main() {
	log.SetOutput(os.Stderr)

	req, err := readFromStdin()
	if err != nil {
		log.Printf("Failed to read message: %v", err)
		os.Exit(1)
	}

	var resp Response
	switch req.Action {
	case "ping":
		resp = Response{OK: true, Data: map[string]string{"status": "ready"}}
	case "detect_agents":
		resp = handleDetect()
	case "install_skills":
		resp = handleInstall(req.Payload)
	default:
		resp = Response{OK: false, Error: "unknown action: " + req.Action}
	}

	if err := writeToStdout(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
		os.Exit(1)
	}
}
