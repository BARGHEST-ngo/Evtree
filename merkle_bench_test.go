package evtree

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func BenchmarkLeafHash(b *testing.B) {
	node := &dirNode{
		name: "",
		files: []FileEntry{
			{
				Path:   "evidence/file.bin",
				Size:   1 << 20,
				Sha256: sha256.Sum256([]byte("benchmark-content")),
			},
		},
		subdirs: make(map[string]*dirNode),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashDir(node)
	}
}

func BenchmarkBuildMerkle100(b *testing.B) {
	benchmarkBuildMerkle(b, 100, false)
}

func BenchmarkBuildMerkle1000(b *testing.B) {
	benchmarkBuildMerkle(b, 1000, false)
}

func BenchmarkBuildMerkleNested100(b *testing.B) {
	benchmarkBuildMerkle(b, 100, true)
}

func benchmarkBuildMerkle(b *testing.B, n int, nested bool) {
	b.Helper()
	entries := make([]FileEntry, n)
	for i := range entries {
		var p string
		if nested {
			p = fmt.Sprintf("dir%d/file%d.bin", i%10, i)
		} else {
			p = fmt.Sprintf("file%d.bin", i)
		}
		entries[i] = FileEntry{
			Path:   p,
			Size:   int64(i * 512),
			Sha256: sha256.Sum256([]byte(p)),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildMerkle(entries)
	}
}
