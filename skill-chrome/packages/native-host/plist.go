package main

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
)

// Minimal Apple plist XML parser to extract CFBundleName / CFBundleDisplayName.

type plistDict struct {
	XMLName xml.Name `xml:"plist"`
	Dict    struct {
		Elements []plistElement `xml:",any"`
	} `xml:"dict"`
}

type plistElement struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func parsePlistDisplayName(appPath string) string {
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return ""
	}

	return extractPlistValue(data, "CFBundleDisplayName", "CFBundleName")
}

func extractPlistValue(data []byte, keys ...string) string {
	// Simple state-machine parser for Apple XML plist format:
	// <dict>
	//   <key>CFBundleName</key>
	//   <string>Navos</string>
	// </dict>
	type plist struct {
		Dict struct {
			Inner []byte `xml:",innerxml"`
		} `xml:"dict"`
	}
	var p plist
	if err := xml.Unmarshal(data, &p); err != nil {
		return ""
	}

	inner := string(p.Dict.Inner)
	for _, key := range keys {
		val := findKeyValue(inner, key)
		if val != "" {
			return val
		}
	}
	return ""
}

func findKeyValue(inner string, targetKey string) string {
	// Parse the raw XML tokens to find <key>targetKey</key><string>value</string>
	decoder := xml.NewDecoder(stringReader(inner))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "key" {
			continue
		}
		var keyVal string
		if err := decoder.DecodeElement(&keyVal, &start); err != nil {
			continue
		}
		if keyVal != targetKey {
			continue
		}
		// Next element should be <string>value</string>
		for {
			tok2, err := decoder.Token()
			if err != nil {
				return ""
			}
			start2, ok := tok2.(xml.StartElement)
			if !ok {
				continue
			}
			if start2.Name.Local == "string" {
				var val string
				if err := decoder.DecodeElement(&val, &start2); err == nil {
					return val
				}
			}
			break
		}
	}
	return ""
}

type stringReaderType struct {
	s string
	i int
}

func stringReader(s string) *stringReaderType {
	return &stringReaderType{s: s}
}

func (r *stringReaderType) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}
