package evtree

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLoggerChain(t *testing.T) {
	var buf bytes.Buffer
	l := AuditLog(&buf)

	first, err := l.Log(ActionOpen, OutcomeSuccess, TrailEntry{Subject: "opened case"})
	if err != nil {
		t.Fatalf("first Log: %v", err)
	}
	if first.Seq != 1 {
		t.Errorf("first.Seq = %d, want 1", first.Seq)
	}
	if first.PrevHash != "" {
		t.Errorf("first.PrevHash = %q, want empty", first.PrevHash)
	}
	if first.EntryHash == "" {
		t.Error("first.EntryHash is empty")
	}

	second, err := l.Log(ActionAcquire, OutcomeSuccess, TrailEntry{Subject: "acquired image"})
	if err != nil {
		t.Fatalf("second Log: %v", err)
	}
	if second.Seq != 2 {
		t.Errorf("second.Seq = %d, want 2", second.Seq)
	}
	if second.PrevHash != first.EntryHash {
		t.Errorf("second.PrevHash = %q, want %q", second.PrevHash, first.EntryHash)
	}

	if got := strings.Count(buf.String(), "\n"); got != 2 {
		t.Errorf("persisted lines = %d, want 2", got)
	}
}

func TestLoggerEntryFields(t *testing.T) {
	var buf bytes.Buffer
	l := AuditLog(&buf)
	in := TrailEntry{
		Examiner:   "Alice",
		Org:        "Forensics Lab",
		CaseNumber: "CASE-001",
		ExhibitRef: "USB-01",
		Subject:    "acquired image",
	}

	got, err := l.Log(ActionAcquire, OutcomeSuccess, in)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}

	if got.Action != ActionAcquire {
		t.Errorf("Action = %q, want %q", got.Action, ActionAcquire)
	}
	if got.Outcome != OutcomeSuccess {
		t.Errorf("Outcome = %q, want %q", got.Outcome, OutcomeSuccess)
	}
	if got.Examiner != in.Examiner {
		t.Errorf("Examiner = %q, want %q", got.Examiner, in.Examiner)
	}
	if got.Org != in.Org {
		t.Errorf("Org = %q, want %q", got.Org, in.Org)
	}
	if got.CaseNumber != in.CaseNumber {
		t.Errorf("CaseNumber = %q, want %q", got.CaseNumber, in.CaseNumber)
	}
	if got.ExhibitRef != in.ExhibitRef {
		t.Errorf("ExhibitRef = %q, want %q", got.ExhibitRef, in.ExhibitRef)
	}
	if got.Subject != in.Subject {
		t.Errorf("Subject = %q, want %q", got.Subject, in.Subject)
	}
	if got.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", got.PID, os.Getpid())
	}
	if got.Timestamp.IsZero() {
		t.Error("Timestamp not set")
	}
	if got.EntryHash == "" {
		t.Error("EntryHash is empty")
	}
}
