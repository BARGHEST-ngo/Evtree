package evtree

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Hash32 [32]byte

func (h Hash32) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash32) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h *Hash32) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := Hash32FromHex(s)
	if err != nil {
		return err
	}
	*h = parsed
	return nil
}

func Hash32FromHex(s string) (Hash32, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Hash32{}, err
	}
	if len(b) != 32 {
		return Hash32{}, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	var h Hash32
	copy(h[:], b)
	return h, nil
}
