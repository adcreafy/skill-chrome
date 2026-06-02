package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleDetect(t *testing.T) {
	resp := handleDetect()
	if !resp.OK {
		t.Fatalf("expected OK=true, got error: %s", resp.Error)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	engines, ok := result["engines"].([]interface{})
	if !ok {
		t.Fatalf("expected engines array, got %T", result["engines"])
	}

	if len(engines) < 3 {
		t.Errorf("expected at least 3 engines, got %d", len(engines))
	}

	engineMap := make(map[string]bool)
	for _, e := range engines {
		eng := e.(map[string]interface{})
		id := eng["id"].(string)
		engineMap[id] = true
	}
	for _, expected := range []string{"cursor", "claude-code", "openclaw", "hermes"} {
		if !engineMap[expected] {
			t.Errorf("missing engine %s in detect result", expected)
		}
	}
}

func TestHandleDetect_DetectsExistingDirs(t *testing.T) {
	home, _ := os.UserHomeDir()

	cursorDir := filepath.Join(home, ".cursor")
	if _, err := os.Stat(cursorDir); err != nil {
		t.Skip("~/.cursor not found, skipping detection test")
	}

	resp := handleDetect()
	data, _ := json.Marshal(resp)

	var result struct {
		Engines []DetectedEngine `json:"engines"`
	}
	json.Unmarshal(data, &result)

	for _, eng := range result.Engines {
		if eng.ID == "cursor" {
			if !eng.Detected {
				t.Error("expected cursor to be detected")
			}
			return
		}
	}
	t.Error("cursor engine not found in results")
}
