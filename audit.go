package evtree

import "time"

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

func compare(comp1 Bag, comp2 Bag) ([]Added, []Deleted, []Modified, error) {
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

	for name, node := range comp1.Root.Children {
		if node.Children != nil {
			continue
		}
		if v, ok := comp2.Root.Children[name]; ok {
			if node.Hash != v.Hash {
				modified = append(modified, Modified{
					Path:      name,
					MtimeOld:  mtime1[name],
					MtimeNew:  mtime2[name],
					Sha256Old: node.Hash,
					Sha256New: v.Hash,
				})
			}
		} else {
			deleted = append(deleted, Deleted{Event{
				Mtime:  mtime1[name],
				Path:   name,
				Sha256: node.Hash,
			}})
		}
	}
	for name, node := range comp2.Root.Children {
		if node.Children != nil {
			continue
		}
		if _, ok := comp1.Root.Children[name]; !ok {
			added = append(added, Added{Event{
				Mtime:  mtime2[name],
				Path:   name,
				Sha256: node.Hash,
			}})
		}
	}
	return added, deleted, modified, nil
}
