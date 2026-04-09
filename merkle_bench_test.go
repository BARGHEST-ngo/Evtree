package evtree

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

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
			Sha256: Hash32(sha256.Sum256([]byte(p))),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildMerkle(entries)
	}
}
