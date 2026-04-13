package evtree

import (
	"encoding/json"
	"errors"
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

type Acquisition struct {
	Timestamp      time.Time    `json:"acquired_at"`
	Case           CaseMetadata `json:"case"`
	Entries        []FileEntry  `json:"entries"`
	Root           *TreeNode    `json:"root"`
	TimestampToken []byte       `json:"timestamp_token,omitempty"`
	TimestampedAt  time.Time    `json:"timestamped_at,omitempty"`
}

func (m CaseMetadata) Validate() error {
	if m.CaseNumber == "" {
		return errors.New("case_number is required")
	}
	if m.ExhibitRef == "" {
		return errors.New("exhibit_ref is required")
	}
	if m.Examiner == "" {
		return errors.New("examiner is required")
	}
	if m.Organisation == "" {
		return errors.New("organisation is required")
	}
	return nil
}

func Acquire(root string, meta CaseMetadata) (Acquisition, []EvidenceError, error) {
	if err := meta.Validate(); err != nil {
		return Acquisition{}, nil, err
	}
	entries, everror, err := AcquireDir(root)
	if err != nil {
		return Acquisition{}, everror, err
	}
	return Acquisition{
		Timestamp: time.Now(),
		Case:      meta,
		Entries:   entries,
		Root:      buildMerkle(entries),
	}, everror, nil
}

func (b Acquisition) Save(filename string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadAcquisition(path string) (Acquisition, error) {
	file, err := os.Open(path)
	if err != nil {
		return Acquisition{}, err
	}
	defer file.Close()
	var b Acquisition
	if err := json.NewDecoder(file).Decode(&b); err != nil {
		return Acquisition{}, err
	}
	return b, nil
}
