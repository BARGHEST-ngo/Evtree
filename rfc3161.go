package evtree

import (
	"bytes"
	"crypto"
	"fmt"
	"io"
	"net/http"

	"github.com/digitorus/timestamp"
)

const DefaultTSA = "https://freetsa.org/tsr"

func Timestamp(bag *Bag, tsaURL string) error {
	hash := bag.Root.Hash

	req, err := timestamp.CreateRequest(bytes.NewReader(hash[:]), &timestamp.RequestOptions{
		Hash:         crypto.SHA256,
		Certificates: true,
	})
	if err != nil {
		return fmt.Errorf("creating timestamp request: %w", err)
	}

	resp, err := http.Post(tsaURL, "application/timestamp-query", bytes.NewReader(req))
	if err != nil {
		return fmt.Errorf("sending timestamp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TSA returned status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading timestamp response: %w", err)
	}

	ts, err := timestamp.ParseResponse(token)
	if err != nil {
		return fmt.Errorf("parsing timestamp response: %w", err)
	}

	if !bytes.Equal(ts.HashedMessage, hash[:]) {
		return fmt.Errorf("TSA response hash does not match bag root hash")
	}

	bag.TimestampToken = token
	bag.TimestampedAt = ts.Time
	return nil
}

func VerifyTimestamp(bag Bag) error {
	if bag.TimestampToken == nil {
		return fmt.Errorf("bag has no timestamp token")
	}

	ts, err := timestamp.ParseResponse(bag.TimestampToken)
	if err != nil {
		return fmt.Errorf("parsing timestamp token: %w", err)
	}

	hash := bag.Root.Hash
	if !bytes.Equal(ts.HashedMessage, hash[:]) {
		return fmt.Errorf("timestamp hash does not match bag root hash")
	}

	return nil
}
