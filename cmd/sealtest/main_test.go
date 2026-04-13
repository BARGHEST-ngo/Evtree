package main

import (
	"fmt"
	"log"
	"os"

	"filippo.io/age"
	evtree "github.com/BARGHEST-ngo/Evtree"
)

func main() {
	evidenceDir := ""
	outPath := "/tmp/evidence.zip.age"

	// Generate a throwaway key pair for testing
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		log.Fatal("keygen:", err)
	}
	recipient := identity.Recipient()
	fmt.Println("Public key:", recipient)

	// Acquire
	meta := evtree.CaseMetadata{
		CaseNumber:   "TEST-001",
		ExhibitRef:   "EX-01",
		Examiner:     "Ovi",
		Organisation: "BARGHEST",
	}
	fmt.Println("Acquiring...")
	acquisition, errs, err := evtree.Acquire(evidenceDir, meta)
	if err != nil {
		log.Fatal("acquire:", err)
	}
	if len(errs) > 0 {
		fmt.Printf("  %d files could not be read\n", len(errs))
	}
	fmt.Printf("  Root hash: %x\n", acquisition.Root.Hash)

	// Seal
	fmt.Println("Sealing to", outPath)
	if err := evtree.Seal(acquisition, evidenceDir, recipient, outPath); err != nil {
		log.Fatal("seal:", err)
	}

	info, _ := os.Stat(outPath)
	fmt.Printf("  Sealed archive: %d bytes\n", info.Size())

	manifestInfo, _ := os.Stat(outPath + ".json")
	fmt.Printf("  Detached manifest: %d bytes\n", manifestInfo.Size())

	// Unseal
	fmt.Println("Unsealing to /tmp/evidence-out.zip")
	recovered, err := evtree.Unseal(outPath, identity, "/tmp/evidence-out.zip")
	if err != nil {
		log.Fatal("unseal:", err)
	}
	fmt.Printf("  Recovered root hash: %x\n", recovered.Root.Hash)

	info2, _ := os.Stat("/tmp/evidence-out.zip")
	fmt.Printf("  Decrypted ZIP: %d bytes\n", info2.Size())

	if recovered.Root.Hash == acquisition.Root.Hash {
		fmt.Println("Root hashes match - seal/unseal OK")
	} else {
		fmt.Println("MISMATCH - something went wrong")
	}
}
