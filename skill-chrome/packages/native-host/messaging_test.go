package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestReadWriteMessage(t *testing.T) {
	req := Request{
		Action:  "ping",
		Payload: nil,
	}
	body, _ := json.Marshal(req)

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(body))); err != nil {
		t.Fatal(err)
	}
	buf.Write(body)

	got, err := readMessage(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != "ping" {
		t.Errorf("expected action=ping, got %s", got.Action)
	}

	resp := Response{OK: true, Data: map[string]string{"status": "ready"}}
	var out bytes.Buffer
	if err := writeMessage(&out, resp); err != nil {
		t.Fatal(err)
	}

	var length uint32
	outReader := bytes.NewReader(out.Bytes())
	if err := binary.Read(outReader, binary.LittleEndian, &length); err != nil {
		t.Fatal(err)
	}
	respBody := make([]byte, length)
	outReader.Read(respBody)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatal(err)
	}
	if result["ok"] != true {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}
	if result["status"] != "ready" {
		t.Errorf("expected status=ready, got %v", result["status"])
	}
}

func TestResponseMarshal_WithData(t *testing.T) {
	resp := Response{
		OK: true,
		Data: map[string]interface{}{
			"engines": []string{"cursor", "claude-code"},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["ok"] != true {
		t.Errorf("expected ok=true")
	}
	engines, ok := m["engines"].([]interface{})
	if !ok || len(engines) != 2 {
		t.Errorf("expected 2 engines, got %v", m["engines"])
	}
}

func TestResponseMarshal_ErrorOnly(t *testing.T) {
	resp := Response{OK: false, Error: "something broke"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	if m["ok"] != false {
		t.Errorf("expected ok=false")
	}
	if m["error"] != "something broke" {
		t.Errorf("expected error message, got %v", m["error"])
	}
}
