package main

import (
	"testing"
)

func TestExtractPlistValue(t *testing.T) {
	plistXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDisplayName</key>
	<string>Navos</string>
	<key>CFBundleName</key>
	<string>NavosInternal</string>
	<key>CFBundleIdentifier</key>
	<string>com.navos.app</string>
</dict>
</plist>`)

	val := extractPlistValue(plistXML, "CFBundleDisplayName", "CFBundleName")
	if val != "Navos" {
		t.Errorf("expected 'Navos', got %q", val)
	}

	val2 := extractPlistValue(plistXML, "CFBundleName")
	if val2 != "NavosInternal" {
		t.Errorf("expected 'NavosInternal', got %q", val2)
	}
}

func TestExtractPlistValue_MissingKey(t *testing.T) {
	plistXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.test.app</string>
</dict>
</plist>`)

	val := extractPlistValue(plistXML, "CFBundleDisplayName", "CFBundleName")
	if val != "" {
		t.Errorf("expected empty string, got %q", val)
	}
}
