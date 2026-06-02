package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
)

// Request is the message sent from the Chrome extension.
type Request struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Response is the message sent back to the Chrome extension.
type Response struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"-"`
}

func (r Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	if r.Data != nil {
		raw, err := json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
		merged := map[string]json.RawMessage{
			"ok": mustMarshal(r.OK),
		}
		if r.Error != "" {
			merged["error"] = mustMarshal(r.Error)
		}
		var dataMap map[string]json.RawMessage
		if err := json.Unmarshal(raw, &dataMap); err == nil {
			for k, v := range dataMap {
				merged[k] = v
			}
		}
		return json.Marshal(merged)
	}
	return json.Marshal(struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}{r.OK, r.Error})
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func readMessage(r io.Reader) (*Request, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	var req Request
	if err := json.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func writeMessage(w io.Writer, resp Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func readFromStdin() (*Request, error) {
	return readMessage(os.Stdin)
}

func writeToStdout(resp Response) error {
	return writeMessage(os.Stdout, resp)
}
