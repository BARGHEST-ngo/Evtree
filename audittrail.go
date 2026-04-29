package evtree

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Action string
const (
	ActionOpen	Action = "open"
	ActionAcquire	Action = "acquire"
	ActionSeal	Action = "seal"
	ActionUnseal	Action = "unseal"
	ActionVerify	Action = "verify"
	ActionTimestamp	Action = "timestamp"
	ActionTransfer	Action = "transfer"
	ActionAccess	Action = "access"
	ActionClose	Action = "close"
)

type Outcome string
const (
	OutcomeSuccess	Outcome = "success"
	OutcomeFailure	Outcome = "failure"
)

type TrailEntry struct {
	Seq		uint64			`json:"seq"`
	Timestamp	time.Time		`json:"timestamp"`
	Action		Action			`json:"action"`
	Outcome		Outcome			`json:"outcome"`
	Examiner	string			`json:"examiner"`
	Org		string			`json:"organisation"`
	Host		string			`json:"host"`
	PID		int			`json:"pid"`
	CaseNumber	string			`json:"case_number"`
	ExhibitRef	string			`json:"exhibit_ref"`
	Subject		string			`json:"subject"`
	Details		map[string]string	`json:"details,omitempty"`
	ErrorMsg	string			`json:"error,omitempty"`
	PrevHash	string			`json:"prev_hash"`
	EntryHash	string			`json:"entry_hash"`
}

type AuditTrail struct {
	mu		sync.Mutex
	path		string
	file		*os.File
	case_		CaseMetadata
	seq		uint64
	lastHash	string
	closed		bool
}

func OpenTrail(path string, meta CaseMetadata, actor string) (*AuditTrail, error) {
	if err := meta.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *AuditTrail) Log(action Action, outcome Outcome, subject string, details map[string]string, err error) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return errors.New("trail closed")
	}

	host, _ := os.Hostname()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	entry := TrailEntry{
		Seq:		t.seq + 1,
		Timestamp:	time.Now().UTC(),
		Action:		action,
		Outcome:	outcome,
		Examiner:	t.case_.Examiner,
		Org:		t.case_.Organisation,
		Host:		host,
		PID:		os.Getpid(),
		CaseNumber:	t.case_.CaseNumber,
		ExhibitRef:	t.case_.ExhibitRef,
		Subject:	subject,
		Details:	details,
		ErrorMsg:	errMsg,
		PrevHash:	t.lastHash,
		EntryHash:	"",
	}

	entry.EntryHash = ""
	payload, mErr := json.Marshal(entry)
	if mErr != nil {
		return fmt.Errorf("canonicalise entry: %w", mErr)
	}
	sum := sha256.Sum256(payload)
	entry.EntryHash = hex.EncodeToString(sum[:])

	line, mErr := json.Marshal(entry)
	if mErr != nil {
		return fmt.Errorf("marshal entry: %w", mErr)
	}
	line = append(line, '\n')

	if _, wErr := t.file.Write(line); wErr != nil {
		return fmt.Errorf("write entry: %w", wErr)
	}
	if sErr := t.file.Sync(); sErr != nil {
		return fmt.Errorf("fsync trail: %w", sErr)
	}

	t.seq = entry.Seq
	t.lastHash = entry.EntryHash
	return nil
}
