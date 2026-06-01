package evtree

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// to keep consistency
type Action string

const (
	ActionOpen      Action = "open"
	ActionAcquire   Action = "acquire"
	ActionSeal      Action = "seal"
	ActionUnseal    Action = "unseal"
	ActionVerify    Action = "verify"
	ActionTimestamp Action = "timestamp"
	ActionTransfer  Action = "transfer"
	ActionAccess    Action = "access"
	ActionClose     Action = "close"
)

// we want to log even failures as evidence
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

type TrailEntry struct {
	Seq        uint64            `json:"seq"`
	Timestamp  time.Time         `json:"timestamp"`
	Action     Action            `json:"action"`
	Outcome    Outcome           `json:"outcome"`
	Examiner   string            `json:"examiner"`
	Org        string            `json:"organisation"`
	Host       string            `json:"host"`
	PID        int               `json:"pid"`
	CaseNumber string            `json:"case_number"`
	ExhibitRef string            `json:"exhibit_ref"`
	Subject    string            `json:"subject"`
	Details    map[string]string `json:"details,omitempty"`
	ErrorMsg   string            `json:"error,omitempty"`
	PrevHash   string            `json:"prev_hash"`
	EntryHash  string            `json:"entry_hash"`
}

type Logger struct {
	mu       sync.Mutex
	w        io.Writer
	seq      uint64
	prevHash string
}

func (l *Logger) Log(action Action, outcome Outcome, input TrailEntry) (TrailEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	input.Seq = l.seq + 1
	input.Timestamp = time.Now().UTC()
	input.Action = action
	input.Outcome = outcome
	input.Host, _ = os.Hostname()
	//Not 100% sure if we need this
	input.PID = os.Getpid()
	input.PrevHash = l.prevHash

	serialize, err := json.Marshal(input)
	if err != nil {
		return TrailEntry{}, err
	}

	hash := Hash32(sha256.Sum256(serialize))

	input.EntryHash = hash.String()
	l.seq++
	l.prevHash = hash.String()

	//remarshal to include entryhash
	record, err := json.Marshal(input)
	if err != nil {
		return TrailEntry{}, err
	}
	if _, err := l.w.Write(append(record, '\n')); err != nil {
		return TrailEntry{}, err
	}

	return input, nil
}

func AuditLog(w io.Writer) *Logger {
	return &Logger{w: w}
}

//TODO verify hashs in log
