package evtree

import (
	"encoding/json"
	"os"
	"time"
)

type CaseMetadata struct {
	CaseNumber   string `json:"case_number"`
	ExhibitRef   string `json:"exhibit_ref"`
	Examiner     string `json:"examiner"`
	Organisation string `json:"organisation"`
	DeviceIMEI   string `json:"device_imei,omitempty"`
	DeviceSerial string `json:"device_serial,omitempty"`
	DeviceModel  string `json:"device_model,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

type Bag struct {
	Timestamp time.Time    `json:"acquired_at"`
	Case      CaseMetadata `json:"case"`
	Entries   []FileEntry  `json:"entries"`
	Root      *TreeNode    `json:"root"`
}

func Acquire(root string, meta CaseMetadata) (Bag, []EvidenceError, error) {
	entries, everror, err := AcquireDir(root)
	if err != nil {
		return Bag{}, everror, err
	}
	return Bag{
		Timestamp: time.Now(),
		Case:      meta,
		Entries:   entries,
		Root:      buildMerkle(entries),
	}, everror, nil
}

func (b Bag) Save(filename string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadBag(path string) (Bag, error) {
	file, err := os.Open(path)
	if err != nil {
		return Bag{}, err
	}
	defer file.Close()
	var b Bag
	if err := json.NewDecoder(file).Decode(&b); err != nil {
		return Bag{}, err
	}
	return b, nil
}
