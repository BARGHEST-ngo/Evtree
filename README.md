# Evtree

Evtree (Evidence Tree) is a Go library for forensic evidence integrity, providing Merkle tree acquisition, chain of custody documentation, structured audit trails, and tamper-evident sealing. At its core, the library computes a deterministic, directory-aware Merkle tree hash over a set of files. Given a list of file entries, each described by a relative path, byte size, SHA-256 digest, and modification time, the library reconstructs the directory hierarchy from the file paths and recursively hashes each directory node. Leaf nodes are constructed by prepending a domain separation byte (0x00) to a canonical string representation of the file metadata, then computing the SHA-256 digest. Internal directory nodes are computed by sorting their children lexicographically by name, concatenating the child hashes with a distinct domain separation byte (0x01), and hashing the result. This structure ensures that the final root hash is sensitive to both the content and the hierarchical organisation of the file tree, and that it is fully deterministic regardless of the order in which file entries are provided. The library is primarily used for projects: [MESH](https://github.com/BARGHEST-ngo/MESH), where it provides tamper-evident integrity verification of acquired forensic artifacts via [androidqf](https://github.com/mvt-project/androidqf).

During acquisition, the library collects file entries into a signed evidence bag alongside case metadata — including case number, exhibit reference, examiner identity, device identifiers, and organisational details — providing a structured record of the circumstances under which the evidence was obtained. Files that cannot be read during acquisition, whether due to access restrictions or device errors, are recorded as evidence errors with a timestamp and the reason for failure rather than causing the acquisition to abort. This ensures that partial acquisitions are documented rather than silently discarded, which is critical when dealing with locked or protected files on seized devices.

This is particularly important for maintaining chain of custody in digital forensic investigations. When evidence is acquired from a device, any subsequent handling, transfer, or storage introduces the possibility of accidental or deliberate modification. A single root hash computed at the time of acquisition serves as a cryptographic seal over the entire evidence set. At any later stage, whether during analysis, peer review, or courtroom presentation, the same hash can be recomputed from the files on hand and compared against the original. If even a single byte in any file has changed, or if a file has been added, removed, or moved to a different directory, the root hash will differ, immediately revealing that the evidence has been altered. Because the tree mirrors the directory structure, it is also possible to isolate which branch of the hierarchy was affected without rehashing the entire collection, identifying precisely which files were added, deleted, or modified between any two acquisitions. This provides both a tamper detection mechanism and an efficient means of auditing evidence integrity across custodial transfers.

This work is inspired by [ECo-Bag: An elastic container based on merkle tree as a universal digital evidence bag](https://www.sciencedirect.com/science/article/abs/pii/S2666281724000404). Acknowledgements to the authors.

## Installation

```
go get evtree
```

## API

| Function | Description |
|---|---|
| `Acquire(root string) (Bag, []EvidenceError, error)` | Walk a directory and produce a signed evidence bag |
| `AcquireDir(root string) ([]FileEntry, []EvidenceError, error)` | Walk a directory and return raw file entries |
| `MerkleFromDir(root string) (Hash32, error)` | Compute the root Merkle hash of a directory |
| `Compare(comp1, comp2 Bag) ([]Added, []Deleted, []Modified, error)` | Compare two bags and return changes |
| `(Bag) Save(filename string) error` | Serialise a bag to JSON |
| `LoadBag(path string) (Bag, error)` | Load a bag from JSON |

> **Planned:** `Verify(bag Bag, root string)` — re-acquire a live directory and compare against a saved bag in one call.

## Usage

```go
// Acquire evidence at time of seizure
bag1, errs, err := evtree.Acquire("/path/to/evidence")
if err != nil {
    log.Fatal(err)
}
if len(errs) > 0 {
    // handle files that could not be read
}

// Save to disk
if err := bag1.Save("evidence.json"); err != nil {
    log.Fatal(err)
}

// Later: reload and compare against a re-acquired directory
bag2, _, err := evtree.Acquire("/path/to/evidence")
if err != nil {
    log.Fatal(err)
}

added, deleted, modified, err := evtree.Compare(bag1, bag2)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Added: %d  Deleted: %d  Modified: %d\n", len(added), len(deleted), len(modified))
```
