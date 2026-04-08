package evtree

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

func AquireDir(root string) ([]FileEntry, error) {
	var entries []FileEntry
	err := godirwalk.Walk(root, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			//since merkle tree reconstructs the directory from the file paths, we can skip dir entries
			if de.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			fi, err := os.Stat(path)
			if err != nil {
				return nil
			}

			sum, err := sha256File(path)
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
		return []FileEntry{}, err
	}
	return entries, nil
}

func MerkleFromDir(root string) ([32]byte, error) {
	entries, err := AquireDir(root) 
	if err != nil {
		return [32]byte{}, err
	}
	return buildMerkle(entries), nil
}

func sha256File(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}

	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}
