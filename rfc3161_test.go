package evtree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyTimestampNoToken(t *testing.T) {
	bag := Acquisition{}
	if err := VerifyTimestamp(bag); err == nil {
		t.Error("expected error for bag with no timestamp token")
	}
}

func TestTimestampInvalidTSA(t *testing.T) {
	bag := makeAcquisition(t)
	err := Timestamp(&bag, "http://invalid.tsa.example.com")
	if err == nil {
		t.Error("expected error for invalid TSA URL")
	}
}

func TestTimestampAndVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RFC 3161 integration test in short mode")
	}

	bag := makeAcquisition(t)

	if err := Timestamp(&bag, DefaultTSA); err != nil {
		t.Fatalf("Timestamp: %v", err)
	}

	if bag.TimestampToken == nil {
		t.Fatal("expected timestamp token to be set")
	}
	if bag.TimestampedAt.IsZero() {
		t.Fatal("expected timestamped_at to be set")
	}

	if err := VerifyTimestamp(bag); err != nil {
		t.Fatalf("VerifyTimestamp: %v", err)
	}
}

func TestVerifyTimestampTamperedAcquisition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RFC 3161 integration test in short mode")
	}

	bag := makeAcquisition(t)

	if err := Timestamp(&bag, DefaultTSA); err != nil {
		t.Fatalf("Timestamp: %v", err)
	}

	// Tamper with the root hash after timestamping
	bag.Root.Hash = Hash32{}

	if err := VerifyTimestamp(bag); err == nil {
		t.Error("expected verification to fail after tampering with root hash")
	}
}

func makeAcquisition(t *testing.T) Acquisition {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "evidence.txt"), []byte("test evidence"), 0644); err != nil {
		t.Fatal(err)
	}
	meta := CaseMetadata{
		CaseNumber:   "TEST-001",
		ExhibitRef:   "EX-01",
		Examiner:     "Test Examiner",
		Organisation: "Test Lab",
	}
	bag, _, err := Acquire(dir, meta)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	return bag
}
