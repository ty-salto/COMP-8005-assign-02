package messages

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type Type string

const (
	REGISTER Type = "REGISTER"
	ACK      Type = "ACK"
	JOB      Type = "JOB"
	RESULT   Type = "RESULT"

	HEARTBEAT_REQ Type = "HEARTBEAT_REQ"
	HEARTBEAT_RES Type = "HEARTBEAT_RES"

)

// HB Controller -> Worker
type HeartbeatReq struct {
	Type Type   `json:"type"`
	Seq  uint64 `json:"seq"`
}

// HB Worker -> Controller
type HeartbeatRes struct {
	Type   Type   `json:"type"`
	Seq    uint64 `json:"seq"`
    Tested uint64 `json:"tested"`
	DeltaTested uint64 `json:"delta_tested"` // candidates tested so far after last 

	Done  bool  `json:"done"`
	Found bool  `json:"found"`
}

// evelope for communicating between controller and worker 
type Envelope struct {
	Type Type `json:"type"`
}

type RegisterMsg struct {
	Type    Type   `json:"type"`
	Worker  string `json:"worker"`
}

type AckMsg struct {
	Type   Type   `json:"type"`
	Status string `json:"status"` // "OK" or "ERROR"
	Error  string `json:"error,omitempty"`
}

type JobMsg struct {
	Type        Type   `json:"type"`
	Username    string `json:"username"`
	FullHash    string `json:"full_hash"`
	Alg         string `json:"alg"`          // "yescrypt"|"bcrypt"|"sha256"|"sha512"|"md5"
	Charset     string `json:"charset"`      // must match required 79-char set
}

type ResultMsg struct {
	Type            Type   `json:"type"`
	Status          string `json:"status"` // "FOUND"|"NOT_FOUND"|"ERROR"
	Password        string `json:"password,omitempty"`
	Error           string `json:"error,omitempty"`
	WorkerComputeNs int64  `json:"worker_compute_ns"`
}

// --- Simple NDJSON helpers (one JSON object per line) ---
func Send(conn net.Conn, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(b, '\n'))
	return err
}

// RecvLine reads one line and unmarshals into out.
func RecvLine(r *bufio.Reader, out any) error {
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return fmt.Errorf("empty message")
	}
	return json.Unmarshal([]byte(line), out)
}

// RecvRawLine reads exactly one newline-delimited JSON message (one line).
func RecvRawLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty message")
	}
	return []byte(line), nil
}

// PeekType extracts the "type" field without knowing the concrete struct yet.
func PeekType(raw []byte) (Type, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", err
	}
	if env.Type == "" {
		return "", fmt.Errorf("missing type field")
	}
	return env.Type, nil
}
