package evtree

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

func AcquireDir(root string) ([]FileEntry, error) {
	var entries []FileEntry
	err := godirwalk.Walk(root, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil {
				return err
			}
			sum, err := sha256Reader(f)
			if err != nil {
				return err
			}

			entries = append(entries, FileEntry{
				Path:   rel,
				Size:   fi.Size(),
				Sha256: sum,
			})
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func MerkleFromDir(root string) ([32]byte, error) {
	entries, err := AcquireDir(root)
	if err != nil {
		return [32]byte{}, err
	}
	return buildMerkle(entries), nil
}

func sha256Reader(r io.Reader) ([32]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return [32]byte{}, err
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}
