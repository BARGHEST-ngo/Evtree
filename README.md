# Evtree

Evtree (Evidence Tree) is a Go library for forensic evidence integrity, providing Merkle tree based integrity validator, chain of custody documentation, RFC 3161 trusted timestamping, structured audit trails, and tamper-evident sealing. 

At its core, the library computes a deterministic, directory-aware Merkle tree hash over a set of files. Given a list of file entries, each described by a relative path, byte size, SHA-256 digest, and modification time, the library reconstructs the directory hierarchy from the file paths and recursively hashes each directory node. Leaf nodes are constructed by prepending a domain separation byte (0x00) to a canonical string representation of the file metadata, then computing the SHA-256 digest. Internal directory nodes are computed by sorting their children lexicographically by name, concatenating the child hashes with a distinct domain separation byte (0x01), and hashing the result. This structure ensures that the final root hash is sensitive to both the content and the hierarchical organisation of the file tree, and that it is fully deterministic regardless of the order in which file entries are provided. 

During acquisition, the library collects file entries into a signed evidence acquisition alongside case metadata — including case number, exhibit reference, examiner identity, device identifiers, and organisational details — providing a structured record of the circumstances under which the evidence was obtained. Files that cannot be read during acquisition, whether due to access restrictions or device errors, are recorded as evidence errors with a timestamp and the reason for failure rather than causing the acquisition to abort. This ensures that partial acquisitions are documented rather than silently discarded, which is critical when dealing with locked or protected files on seized devices.

Once an evidence acquisition has been produced, the library supports RFC 3161 trusted timestamping. The root hash is submitted to a trusted timestamping authority (TSA), which returns a cryptographically signed token binding the hash to a specific point in time. This token is stored within the acquisition and can be verified at any stage by recomputing the root hash and comparing it against the hash embedded in the token. Because the token is signed by an independent third party, it directly addresses the weakness of relying on system clocks — which can be manipulated — by anchoring the acquisition to an externally verifiable time source.

This is particularly important for maintaining chain of custody in digital forensic investigations. When evidence is acquired from a device, any subsequent handling, transfer, or storage introduces the possibility of accidental or deliberate modification. A single root hash computed at the time of acquisition serves as a cryptographic seal over the entire evidence set. At any later stage, whether during analysis, peer review, or courtroom presentation, the same hash can be recomputed from the files on hand and compared against the original. If even a single byte in any file has changed, or if a file has been added, removed, or moved to a different directory, the root hash will differ, immediately revealing that the evidence has been altered. Because the tree mirrors the directory structure, it is also possible to isolate which branch of the hierarchy was affected without rehashing the entire collection, identifying precisely which files were added, deleted, or modified between any two acquisitions. This provides both a tamper detection mechanism and an efficient means of auditing evidence integrity across custodial transfers.

The library is primarily used for [MESH](https://github.com/BARGHEST-ngo/MESH), where it provides tamper-evident integrity verification of acquired forensic artifacts via [androidqf](https://github.com/mvt-project/androidqf).

This work is inspired by [ECo-Bag: An elastic container based on merkle tree as a universal digital evidence acquisition](https://www.sciencedirect.com/science/article/abs/pii/S2666281724000404). Acknowledgements to the authors and Korea Univ.

## Installation

```
go get github.com/BARGHEST-ngo/Evtree
```

## Reference

### Acquisition

| Function | Description |
|---|---|
| `Acquire(root string, meta CaseMetadata) (Acquisition, []EvidenceError, error)` | Walk a directory and produce an evidence acquisition |
| `AcquireDir(root string) ([]FileEntry, []EvidenceError, error)` | Walk a directory and return raw file entries |
| `MerkleFromDir(root string) (Hash32, error)` | Compute the root Merkle hash of a directory |

### Integrity

| Function | Description |
|---|---|
| `Compare(comp1, comp2 Acquisition) ([]Added, []Deleted, []Modified, error)` | Compare two acquisitions and return changes |
| `Verify(root string, meta CaseMetadata, comp1 string) (Result, []EvidenceError, error)` | Re-acquire a live directory and compare against a saved acquisition in one call |
| `Timestamp(acquisition *Acquisition, tsaURL string) error` | Request an RFC 3161 timestamp from a TSA and store it in the acquisition |
| `VerifyTimestamp(acquisition Acquisition) error` | Verify the stored RFC 3161 timestamp token against the acquisition root hash |

### Storage

| Function | Description |
|---|---|
| `(Acquisition) Save(filename string) error` | Serialise an acquisition to JSON |
| `LoadAcquisition(path string) (Acquisition, error)` | Load an acquisition from JSON |
| `Seal(acquisition Acquisition, evidenceDir string, recipient age.Recipient, outPath string) error` | Encrypt the evidence directory into a tamper-evident age-encrypted ZIP archive, with the acquisition manifest written inside and as a detached JSON file |
| `Unseal(sealedPath string, identity age.Identity, outDir string) (Acquisition, error)` | Decrypt a sealed archive and extract its contents, returning the acquisition manifest |

## TODO

- Audit trail API — structured, appendable log of acquisition, transfer, comparison, and verification events
- Digital signatures — sign the root hash with the examiner's private key for non-repudiation

## Usage

```go
meta := evtree.CaseMetadata{
    CaseNumber:   "2024-001",
    ExhibitRef:   "EX-01",
    Examiner:     "J. Smith",
    Organisation: "Digital Forensics Lab",
    DeviceModel:  "Pixel 7",
}

// Acquire evidence at time of seizure
acquisition, errs, err := evtree.Acquire("/path/to/evidence", meta)
if err != nil {
    log.Fatal(err)
}
if len(errs) > 0 {
    // handle files that could not be read
}

// Anchor to a trusted timestamp
if err := evtree.Timestamp(&acquisition, evtree.DefaultTSA); err != nil {
    log.Fatal(err)
}

// Save to disk
if err := acquisition.Save("evidence.json"); err != nil {
    log.Fatal(err)
}

// Later: reload, verify timestamp, and compare against a re-acquired directory
acquisition1, err := evtree.LoadAcquisition("evidence.json")
if err != nil {
    log.Fatal(err)
}
if err := evtree.VerifyTimestamp(acquisition1); err != nil {
    log.Fatal(err)
}

acquisition2, _, err := evtree.Acquire("/path/to/evidence", meta)
if err != nil {
    log.Fatal(err)
}

added, deleted, modified, err := evtree.Compare(acquisition1, acquisition2)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Added: %d  Deleted: %d  Modified: %d\n", len(added), len(deleted), len(modified))
```
