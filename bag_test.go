package evtree

import (
	"testing"
)

func validMeta() CaseMetadata {
	return CaseMetadata{
		CaseNumber:   "C-001",
		ExhibitRef:   "EX-001",
		Examiner:     "ovi",
		Organisation: "BARGHEST",
	}
}

func TestAcquireEntries_OK(t *testing.T) {
	shaA := testSHA("content-a")
	shaB := testSHA("content-b")
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: shaA},
		{Path: "b.txt", Size: 2, Sha256: shaB},
	}

	acq, err := AcquireEntries(entries, validMeta())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(acq.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(acq.Entries))
	}
	if acq.Case != validMeta() {
		t.Errorf("case metadata not preserved: %+v", acq.Case)
	}
	if acq.Root == nil {
		t.Fatalf("expected non-nil root")
	}

	want := dirHash(leafHash("a.txt", 1, shaA), leafHash("b.txt", 2, shaB))
	if acq.Root.Hash != want {
		t.Errorf("root mismatch: got %x, want %x", acq.Root.Hash, want)
	}
}

func TestAcquireEntries_RootMatchesBuildMerkle(t *testing.T) {
	entries := []FileEntry{
		{Path: "c.txt", Size: 30, Sha256: testSHA("c")},
		{Path: "a.txt", Size: 10, Sha256: testSHA("a")},
		{Path: "b.txt", Size: 20, Sha256: testSHA("b")},
	}

	acq, err := AcquireEntries(entries, validMeta())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acq.Root.Hash != BuildMerkle(entries).Hash {
		t.Errorf("AcquireEntries root must equal BuildMerkle root")
	}
}

func TestAcquireEntries_EmptyEntries(t *testing.T) {
	acq, err := AcquireEntries(nil, validMeta())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := BuildMerkle(nil).Hash
	if acq.Root.Hash != want {
		t.Errorf("empty entries root: got %x, want %x", acq.Root.Hash, want)
	}
}

func TestAcquireEntries_InvalidMeta(t *testing.T) {
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: testSHA("a")},
	}

	if _, err := AcquireEntries(entries, CaseMetadata{}); err == nil {
		t.Errorf("expected error for invalid metadata")
	}
}
