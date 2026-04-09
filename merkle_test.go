package evtree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func testSHA(s string) [32]byte { return sha256.Sum256([]byte(s)) }

func leafHash(name string, size int64, h [32]byte) [32]byte {
	leafStr := fmt.Sprintf("evtree:v1:%s:%d:%s\n", name, size, hex.EncodeToString(h[:]))
	return sha256.Sum256(append([]byte{0x00}, []byte(leafStr)...))
}

func dirHash(hashes ...[32]byte) [32]byte {
	buf := make([]byte, 1, 1+32*len(hashes))
	buf[0] = 0x01
	for _, h := range hashes {
		buf = append(buf, h[:]...)
	}
	return sha256.Sum256(buf)
}

func TestMerkleDeterminism(t *testing.T) {
	entries1 := []FileEntry{
		{Path: "a.txt", Size: 10, Sha256: testSHA("a")},
		{Path: "b.txt", Size: 20, Sha256: testSHA("b")},
		{Path: "c.txt", Size: 30, Sha256: testSHA("c")},
	}
	entries2 := []FileEntry{
		{Path: "c.txt", Size: 30, Sha256: testSHA("c")},
		{Path: "a.txt", Size: 10, Sha256: testSHA("a")},
		{Path: "b.txt", Size: 20, Sha256: testSHA("b")},
	}

	root1 := buildMerkle(entries1)
	root2 := buildMerkle(entries2)

	if root1 != root2 {
		t.Errorf("root must be identical regardless of input order:\n root1=%x\n root2=%x", root1, root2)
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
	sha := testSHA("hello-content")
	e := FileEntry{Path: "hello.txt", Size: 5, Sha256: sha}
	root := buildMerkle([]FileEntry{e})

	want := dirHash(leafHash("hello.txt", 5, sha))
	if root != want {
		t.Errorf("single entry: got %x, want %x", root, want)
	}
}

func TestBuildMerkleTwoEntries(t *testing.T) {
	shaA := testSHA("content-a")
	shaB := testSHA("content-b")
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: shaA},
		{Path: "b.txt", Size: 2, Sha256: shaB},
	}
	root := buildMerkle(entries)

	want := dirHash(leafHash("a.txt", 1, shaA), leafHash("b.txt", 2, shaB))
	if root != want {
		t.Errorf("two entries: got %x, want %x", root, want)
	}
}

func TestBuildMerkleOddEntries(t *testing.T) {
	shaA := testSHA("a")
	shaB := testSHA("b")
	shaC := testSHA("c")
	entries := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: shaA},
		{Path: "b.txt", Size: 2, Sha256: shaB},
		{Path: "c.txt", Size: 3, Sha256: shaC},
	}
	root := buildMerkle(entries)

	if root == ([32]byte{}) {
		t.Fatal("root must not be zero")
	}

	modified := []FileEntry{
		{Path: "a.txt", Size: 1, Sha256: shaA},
		{Path: "b.txt", Size: 2, Sha256: shaB},
		{Path: "c.txt", Size: 3, Sha256: testSHA("c-tampered")},
	}
	rootMod := buildMerkle(modified)

	if root == rootMod {
		t.Error("changing entry data must change the root")
	}
}

func TestBuildMerkleSubdirectories(t *testing.T) {
	shaMain := testSHA("main")
	shaRun := testSHA("run")
	shaHelp := testSHA("help")
	entries := []FileEntry{
		{Path: "main.go", Size: 100, Sha256: shaMain},
		{Path: "cmd/run.go", Size: 50, Sha256: shaRun},
		{Path: "cmd/help.go", Size: 30, Sha256: shaHelp},
	}
	root := buildMerkle(entries)

	hCmd := dirHash(leafHash("help.go", 30, shaHelp), leafHash("run.go", 50, shaRun))
	want := dirHash(hCmd, leafHash("main.go", 100, shaMain))

	if root != want {
		t.Errorf("subdirectories: got %x, want %x", root, want)
	}
}

func TestBuildMerklePathNormalization(t *testing.T) {
	sha := testSHA("f")
	r1 := buildMerkle([]FileEntry{{Path: "dir/file.txt", Size: 10, Sha256: sha}})
	r2 := buildMerkle([]FileEntry{{Path: "dir//file.txt", Size: 10, Sha256: sha}})
	r3 := buildMerkle([]FileEntry{{Path: "/dir/file.txt", Size: 10, Sha256: sha}})

	if r1 != r2 || r1 != r3 {
		t.Errorf("normalized paths should produce same root:\n r1=%x\n r2=%x\n r3=%x", r1, r2, r3)
	}
}

func TestBuildMerkleSizeMatters(t *testing.T) {
	sha := testSHA("same-content")
	r1 := buildMerkle([]FileEntry{{Path: "f.txt", Size: 1, Sha256: sha}})
	r2 := buildMerkle([]FileEntry{{Path: "f.txt", Size: 2, Sha256: sha}})

	if r1 == r2 {
		t.Error("different sizes must produce different roots")
	}
}

func TestBuildMerklePathMatters(t *testing.T) {
	sha := testSHA("same-content")
	r1 := buildMerkle([]FileEntry{{Path: "a.txt", Size: 1, Sha256: sha}})
	r2 := buildMerkle([]FileEntry{{Path: "b.txt", Size: 1, Sha256: sha}})

	if r1 == r2 {
		t.Error("different paths must produce different roots")
	}
}
