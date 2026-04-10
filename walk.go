package evtree

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"time"
	"github.com/karrick/godirwalk"
)

type EvidenceError struct {
	Timestamp time.Time
	Everror error 
	File string
}

func AcquireDir(root string) ([]FileEntry, []EvidenceError, error) {
	var entries []FileEntry
	var evidenceerror []EvidenceError
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
				evidenceerror = append(evidenceerror, EvidenceError{
					Timestamp: time.Now(),
					Everror: err,
					File: rel,
				})
				return nil
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil {
				evidenceerror = append(evidenceerror, EvidenceError{
					Timestamp: time.Now(),
					Everror: err,
					File: rel,
				})
				return nil
			}
			sum, err := sha256Reader(f)
			if err != nil {
				evidenceerror = append(evidenceerror, EvidenceError{
					Timestamp: time.Now(),
					Everror: err,
					File: rel,
				})
				return nil
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
		return nil, nil, err
	}
	return entries, evidenceerror, nil
}

func MerkleFromDir(root string) (Hash32, error) {
	entries, _, err := AcquireDir(root)
	if err != nil {
		return Hash32{}, err
	}
	return buildMerkle(entries).Hash, nil
}

func sha256Reader(r io.Reader) (Hash32, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return Hash32{}, err
	}
	var out Hash32
	copy(out[:], h.Sum(nil))
	return out, nil
}
