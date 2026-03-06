package evtree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Path   string   `json:"path"`
	Size   int64    `json:"size"`
	Sha256 [32]byte `json:"-"`
}

type dirNode struct {
	name    string
	files   []FileEntry
	subdirs map[string]*dirNode
}

func buildMerkle(entries []FileEntry) [32]byte {
	root := buildTree(entries)
	return hashDir(root)
}

func buildTree(entries []FileEntry) *dirNode {
	root := &dirNode{name: "", subdirs: make(map[string]*dirNode)}

	for _, e := range entries {
		p := normalize(e.Path)
		parts := strings.Split(p, "/")

		cur := root
		for _, dir := range parts[:len(parts)-1] {
			if _, ok := cur.subdirs[dir]; !ok {
				cur.subdirs[dir] = &dirNode{name: dir, subdirs: make(map[string]*dirNode)}
			}
			cur = cur.subdirs[dir]
		}

		file := e
		file.Path = parts[len(parts)-1]
		cur.files = append(cur.files, file)
	}

	return root
}

func hashDir(node *dirNode) [32]byte {
	type child struct {
		name string
		hash [32]byte
	}

	var children []child

	for _, f := range node.files {
		leafStr := fmt.Sprintf("evtree:v1:%s:%d:%s\n", f.Path, f.Size, hex.EncodeToString(f.Sha256[:]))
		leafBytes := append([]byte{0x00}, []byte(leafStr)...)
		children = append(children, child{name: f.Path, hash: sha256.Sum256(leafBytes)})
	}

	for name, sub := range node.subdirs {
		children = append(children, child{name: name, hash: hashDir(sub)})
	}

	sort.Slice(children, func(i, j int) bool {
		return children[i].name < children[j].name
	})

	if len(children) == 0 {
		return sha256.Sum256([]byte{0x00})
	}

	combined := []byte{0x01}
	for _, c := range children {
		combined = append(combined, c.hash[:]...)
	}
	return sha256.Sum256(combined)
}

func normalize(p string) string {
	p = filepath.ToSlash(p)
	p = filepath.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}
