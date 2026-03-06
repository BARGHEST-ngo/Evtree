package evtree

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestMerkleDeterminism(t *testing.T) {
	entries1 := []FileEntry{
		{Path: "a.txt", Size: 10, Sha256: "aaaa"},
		{Path: "b.txt", Size: 20, Sha256: "bbbb"},
		{Path: "c.txt", Size: 30, Sha256: "cccc"},
	}

	entries2 := []FileEntry{
		{Path: "c.txt", Size: 30, Sha256: "cccc"},
		{Path: "a.txt", Size: 10, Sha256: "aaaa"},
		{Path: "b.txt", Size: 20, Sha256: "bbbb"},
	}

	root1 := buildMerkle(entries1)
	root2 := buildMerkle(entries2)

	t.Logf("Root1: %x", root1)
	t.Logf("Root2: %x", root2)

	if root1 != root2 {
		t.Errorf("Roots should be identical regardless of input order.\nRoot1: %x\nRoot2: %x", root1, root2)
	}
}

func TestBuildMerkleEmpty(t *testing.T) {
	root := buildMerkle(nil)
	want := sha256.Sum256([]byte{0x00})
	if root != want {
		t.Errorf("empty tree: got %x, want %x", root, want)
	}
}

func TestBuildMerkleSingleEntry(t *testing.T) {
	e := FileEntry{Path: "hello.txt", Size: 5, Sha256: "abcd1234"}
	root := buildMerkle([]FileEntry{e})

	// Leaf hash for the file
	leafStr := fmt.Sprintf("evtree:v1:%s:%d:%s\n", "hello.txt", e.Size, e.Sha256)
	leafHash := sha256.Sum256(append([]byte{0x00}, []byte(leafStr)...))
	// Root directory wraps the single leaf: SHA256(0x01 || leafHash)
	want := sha256.Sum256(append([]byte{0x01}, leafHash[:]...))

	t.Logf("root: %x", root)
	t.Logf("want: %x", want)

	if root != want {
		t.Errorf("single entry: got %x, want %x", root, want)
	}
}

func TestBuildMerkleTwoEntries(t *testing.T) {
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: "aa"},
		{Path: "b.txt", Size: 2, Sha256: "bb"},
	}
	root := buildMerkle(entries)


	leafA := fmt.Sprintf("evtree:v1:%s:%d:%s\n", "a.txt", int64(1), "aa")
	leafB := fmt.Sprintf("evtree:v1:%s:%d:%s\n", "b.txt", int64(2), "bb")
	hA := sha256.Sum256(append([]byte{0x00}, []byte(leafA)...))
	hB := sha256.Sum256(append([]byte{0x00}, []byte(leafB)...))
	// Root dir: SHA256(0x01 || hA || hB).
	combined := append([]byte{0x01}, hA[:]...)
	combined = append(combined, hB[:]...)
	want := sha256.Sum256(combined)

	if root != want {
		t.Errorf("two entries: got %x, want %x", root, want)
	}
}

func TestBuildMerkleOddEntries(t *testing.T) {
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: "aa"},
		{Path: "b.txt", Size: 2, Sha256: "bb"},
		{Path: "c.txt", Size: 3, Sha256: "cc"},
	}
	root := buildMerkle(entries)

	t.Logf("root:    %x", root)

	if root == [32]byte{} {
		t.Error("odd entries: root should not be zero")
	}

	// Changing any entry must change the root.
	modified := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: "aa"},
		{Path: "b.txt", Size: 2, Sha256: "bb"},
		{Path: "c.txt", Size: 3, Sha256: "XX"},
	}
	rootMod := buildMerkle(modified)
	t.Logf("rootMod: %x", rootMod)

	if root == rootMod {
		t.Error("odd entries: changing data should change root")
	}
}

func TestBuildMerkleSubdirectories(t *testing.T) {
	entries := []FileEntry{
		{Path: "main.go", Size: 100, Sha256: "aaa"},
		{Path: "cmd/run.go", Size: 50, Sha256: "bbb"},
		{Path: "cmd/help.go", Size: 30, Sha256: "ccc"},
	}
	root := buildMerkle(entries)
	t.Logf("root: %x", root)

	// Manually compute: cmd/ dir has help.go and run.go (sorted).
	hHelp := sha256.Sum256(append([]byte{0x00}, []byte(fmt.Sprintf("evtree:v1:%s:%d:%s\n", "help.go", int64(30), "ccc"))...))
	hRun := sha256.Sum256(append([]byte{0x00}, []byte(fmt.Sprintf("evtree:v1:%s:%d:%s\n", "run.go", int64(50), "bbb"))...))
	cmdDir := append([]byte{0x01}, hHelp[:]...)
	cmdDir = append(cmdDir, hRun[:]...)
	hCmd := sha256.Sum256(cmdDir)

	// Root dir has: cmd/ (dir) and main.go (file), sorted: cmd < main.go.
	hMain := sha256.Sum256(append([]byte{0x00}, []byte(fmt.Sprintf("evtree:v1:%s:%d:%s\n", "main.go", int64(100), "aaa"))...))
	rootDir := append([]byte{0x01}, hCmd[:]...)
	rootDir = append(rootDir, hMain[:]...)
	want := sha256.Sum256(rootDir)

	t.Logf("want: %x", want)

	if root != want {
		t.Errorf("subdirectories: got %x, want %x", root, want)
	}
}

func TestBuildMerklePathNormalization(t *testing.T) {
	e1 := []FileEntry{{Path: "dir/file.txt", Size: 10, Sha256: "ff"}}
	e2 := []FileEntry{{Path: "dir//file.txt", Size: 10, Sha256: "ff"}}
	e3 := []FileEntry{{Path: "/dir/file.txt", Size: 10, Sha256: "ff"}}

	r1 := buildMerkle(e1)
	r2 := buildMerkle(e2)
	r3 := buildMerkle(e3)

	t.Logf("r1 (dir/file.txt):  %x", r1)
	t.Logf("r2 (dir//file.txt): %x", r2)
	t.Logf("r3 (/dir/file.txt): %x", r3)

	if r1 != r2 || r1 != r3 {
		t.Errorf("normalized paths should produce same root:\n r1=%x\n r2=%x\n r3=%x", r1, r2, r3)
	}
}
