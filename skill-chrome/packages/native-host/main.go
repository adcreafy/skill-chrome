package main

import (
	"log"
	"os"
)

func main() {
	// Disable log output to stdout (Chrome reads stdout for responses)
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
