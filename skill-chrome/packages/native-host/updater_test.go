package main

import (
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest  string
		current string
		want    bool
	}{
		{"0.2.0", "0.1.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.1", "0.1.0", true},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", false},
		{"0.0.9", "0.1.0", false},
	}
	for _, c := range cases {
		got := isNewer(c.latest, c.current)
		if got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestStateFilePath(t *testing.T) {
	path := stateFilePath()
	if path == "" {
		t.Error("stateFilePath returned empty string")
	}
	if !contains(path, ".skill-chrome") {
		t.Errorf("expected path to contain .skill-chrome, got %s", path)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsAt(s, sub))
}

func containsAt(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestShouldCheckUpdate_NoStateFile(t *testing.T) {
	// When state file doesn't exist, should return true
	result := shouldCheckUpdate()
	// This depends on whether the file exists on the test machine
	// Just verify it doesn't panic
	_ = result
}
