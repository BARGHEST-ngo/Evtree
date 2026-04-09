package evtree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireDirBasic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := AcquireDir(dir)
	if err != nil {
		t.Fatalf("AcquireDir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	byPath := make(map[string]FileEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	if _, ok := byPath["a.txt"]; !ok {
		t.Error("a.txt missing")
	}
	if _, ok := byPath["b.txt"]; !ok {
		t.Error("b.txt missing")
	}
	if byPath["a.txt"].Size != 5 {
		t.Errorf("a.txt size: got %d, want 5", byPath["a.txt"].Size)
	}
	want, _ := sha256Reader(mustOpen(t, filepath.Join(dir, "a.txt")))
	if byPath["a.txt"].Sha256 != want {
		t.Errorf("a.txt sha256 mismatch")
	}
}

func TestAcquireDirNestedDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := AcquireDir(dir)
	if err != nil {
		t.Fatalf("AcquireDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "sub/file.txt" {
		t.Errorf("path: got %q, want %q", entries[0].Path, "sub/file.txt")
	}
}

func TestAcquireDirUnreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: chmod 000 has no effect")
	}
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secret, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secret, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(secret, 0644) })

	_, err := AcquireDir(dir)
	if err == nil {
		t.Error("expected error for unreadable file, got nil")
	}
}

func TestAcquireDirSymlink(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realPath, []byte("evidence data"), 0644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(dir, "link.txt")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skip("symlinks not supported on this platform:", err)
	}

	entries, err := AcquireDir(dir)
	if err != nil {
		t.Fatalf("AcquireDir: %v", err)
	}

	byPath := make(map[string]FileEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	if _, ok := byPath["real.txt"]; !ok {
		t.Error("real.txt missing from entries")
	}
	if _, ok := byPath["link.txt"]; !ok {
		t.Error("link.txt missing — symlinks should be followed and included")
	}
	if byPath["real.txt"].Sha256 != byPath["link.txt"].Sha256 {
		t.Error("real.txt and link.txt should have the same SHA256 (same target content)")
	}
	if byPath["real.txt"].Size != byPath["link.txt"].Size {
		t.Error("real.txt and link.txt should have the same size")
	}
}

func TestMerkleFromDirDeterministic(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt", "gamma.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	r1, err := MerkleFromDir(dir)
	if err != nil {
		t.Fatalf("first MerkleFromDir: %v", err)
	}
	r2, err := MerkleFromDir(dir)
	if err != nil {
		t.Fatalf("second MerkleFromDir: %v", err)
	}
	if r1 != r2 {
		t.Errorf("MerkleFromDir not deterministic: %x != %x", r1, r2)
	}
}

func mustOpen(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}
