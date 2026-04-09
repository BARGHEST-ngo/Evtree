package evtree

import (
	"encoding/json"
	"os"
	"time"
)

type Bag struct {
	Timestamp time.Time `json:"acquired_at"`
	Root      *TreeNode `json:"root"`
}

func Acquire(root string) (Bag, error) {
	entries, err := AcquireDir(root)
	if err != nil {
		return Bag{}, err
	}
	return Bag{
		Timestamp: time.Now(),
		Root:      buildMerkle(entries),
	}, nil
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
