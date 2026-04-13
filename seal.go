package evtree

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"
)

func Seal(acquisition Acquisition, evidenceDir string, recipient age.Recipient, outPath string) error {
	if err := writeDetachedManifest(acquisition, outPath+".json"); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	ageWriter, err := age.Encrypt(f, recipient)
	if err != nil {
		return err
	}
	defer ageWriter.Close()

	zipWriter := zip.NewWriter(ageWriter)
	defer zipWriter.Close()

	if err := addManifestToZip(zipWriter, acquisition); err != nil {
		return err
	}

	return addDirToZip(zipWriter, evidenceDir)
}


// Unseal decrypts an age-encrypted archive produced by Seal, writing the
// decrypted ZIP to outZipPath and returning the acquisition manifest stored inside.
func Unseal(sealedPath string, identity age.Identity, outZipPath string) (Acquisition, error) {
	f, err := os.Open(sealedPath)
	if err != nil {
		return Acquisition{}, err
	}
	defer f.Close()

	ageReader, err := age.Decrypt(f, identity)
	if err != nil {
		return Acquisition{}, err
	}

	zipBytes, err := io.ReadAll(ageReader)
	if err != nil {
		return Acquisition{}, err
	}

	if err := os.WriteFile(outZipPath, zipBytes, 0400); err != nil {
		return Acquisition{}, err
	}

	zr, err := zip.NewReader(bytesReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return Acquisition{}, err
	}

	var acquisition Acquisition
	for _, zf := range zr.File {
		if zf.Name == "acquisition.json" {
			rc, err := zf.Open()
			if err != nil {
				return Acquisition{}, err
			}
			err = json.NewDecoder(rc).Decode(&acquisition)
			rc.Close()
			if err != nil {
				return Acquisition{}, err
			}
			return acquisition, nil
		}
	}

	return Acquisition{}, fmt.Errorf("acquisition.json not found in sealed archive")
}

func writeDetachedManifest(acquisition Acquisition, path string) error {
	data, err := json.MarshalIndent(acquisition, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0400)
}

func addManifestToZip(zw *zip.Writer, acquisition Acquisition) error {
	w, err := zw.Create("acquisition.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(acquisition)
}

func addDirToZip(zw *zip.Writer, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		w, err := zw.Create(rel)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	})
}


type bytesReader []byte

func (b bytesReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b)) {
		return 0, io.EOF
	}
	n := copy(p, b[off:])
	return n, nil
}
