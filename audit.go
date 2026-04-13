package evtree

import (
	"path"
	"time"
)

type Event struct {
	Mtime  time.Time
	Path   string
	Sha256 Hash32
}

type Added struct {
	Event
}

type Deleted struct {
	Event
}

type Modified struct {
	Path      string
	MtimeOld  time.Time
	MtimeNew  time.Time
	Sha256Old Hash32
	Sha256New Hash32
}

type Result struct {
	Verified bool
	Event []Event
	Modified []Modified
}
	

func Verify(root string, meta CaseMetadata, comp1 string) (Result, []EvidenceError, error) {
	acq1, everror, err := Acquire(root, meta)
	if err != nil {
		return Result{}, everror, err
	}

	comp1Load, err := LoadAcquisition(comp1)
	if err != nil {
		return Result{}, everror, err
	}

	add, del, mod, err := Compare(acq1, comp1Load)
	if err != nil {
		return Result{}, everror, err
	}

	if len(add) == 0 && len(del) == 0 && len(mod) == 0 {
		return Result{Verified: true}, everror, nil
	}

	events := make([]Event, 0, len(add)+len(del))
	for _, a := range add {
		events = append(events, a.Event)
	}
	for _, d := range del {
		events = append(events, d.Event)
	}

	return Result{
		Verified: false,
		Event:    events,
		Modified: mod,
	}, everror, nil
}

func Compare(comp1 Acquisition, comp2 Acquisition) ([]Added, []Deleted, []Modified, error) {
	if comp1.Root.Hash == comp2.Root.Hash {
		return nil, nil, nil, nil
	}

	mtime1 := make(map[string]time.Time, len(comp1.Entries))
	for _, e := range comp1.Entries {
		mtime1[e.Path] = e.Mtime
	}
	mtime2 := make(map[string]time.Time, len(comp2.Entries))
	for _, e := range comp2.Entries {
		mtime2[e.Path] = e.Mtime
	}

	var added []Added
	var deleted []Deleted
	var modified []Modified

	walkNodes(comp1.Root, comp2.Root, "", mtime1, mtime2, &added, &deleted, &modified)

	return added, deleted, modified, nil
}

func walkNodes(
	node1, node2 *TreeNode,
	dir string,
	mtime1, mtime2 map[string]time.Time,
	added *[]Added,
	deleted *[]Deleted,
	modified *[]Modified,
) {
	for name, child1 := range node1.Children {
		fullPath := path.Join(dir, name)
		child2, exists := node2.Children[name]

		if child1.Children != nil {
			if !exists || child2.Children == nil {
				collectDeleted(child1, fullPath, mtime1, deleted)
			} else if child1.Hash != child2.Hash {
				walkNodes(child1, child2, fullPath, mtime1, mtime2, added, deleted, modified)
			}
		} else {
			if !exists {
				*deleted = append(*deleted, Deleted{Event{
					Mtime:  mtime1[fullPath],
					Path:   fullPath,
					Sha256: child1.Hash,
				}})
			} else if child1.Hash != child2.Hash {
				*modified = append(*modified, Modified{
					Path:      fullPath,
					MtimeOld:  mtime1[fullPath],
					MtimeNew:  mtime2[fullPath],
					Sha256Old: child1.Hash,
					Sha256New: child2.Hash,
				})
			}
		}
	}

	for name, child2 := range node2.Children {
		fullPath := path.Join(dir, name)
		if _, exists := node1.Children[name]; !exists {
			if child2.Children != nil {
				collectAdded(child2, fullPath, mtime2, added)
			} else {
				*added = append(*added, Added{Event{
					Mtime:  mtime2[fullPath],
					Path:   fullPath,
					Sha256: child2.Hash,
				}})
			}
		}
	}
}

func collectDeleted(node *TreeNode, dir string, mtime1 map[string]time.Time, deleted *[]Deleted) {
	for name, child := range node.Children {
		fullPath := path.Join(dir, name)
		if child.Children != nil {
			collectDeleted(child, fullPath, mtime1, deleted)
		} else {
			*deleted = append(*deleted, Deleted{Event{
				Mtime:  mtime1[fullPath],
				Path:   fullPath,
				Sha256: child.Hash,
			}})
		}
	}
}

func collectAdded(node *TreeNode, dir string, mtime2 map[string]time.Time, added *[]Added) {
	for name, child := range node.Children {
		fullPath := path.Join(dir, name)
		if child.Children != nil {
			collectAdded(child, fullPath, mtime2, added)
		} else {
			*added = append(*added, Added{Event{
				Mtime:  mtime2[fullPath],
				Path:   fullPath,
				Sha256: child.Hash,
			}})
		}
	}
}
