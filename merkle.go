package evtree

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TreeNode struct {
	Hash     Hash32               `json:"merkle_hash"`
	Size     int64                `json:"size,omitempty"`
	Sha256   *Hash32              `json:"sha256,omitempty"`
	Children map[string]*TreeNode `json:"children,omitempty"`
}

type FileEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Sha256 Hash32 `json:"sha256"`
	Mtime time.Time `json:"modificationtime"`
}

type dirNode struct {
	name    string
	files   []FileEntry
	subdirs map[string]*dirNode
}

func buildMerkle(entries []FileEntry) *TreeNode {
	root := buildTree(entries)
	return buildTreeNode(root)
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

func buildTreeNode(node *dirNode) *TreeNode {
	type child struct {
		name string
		hash Hash32
	}

	result := &TreeNode{
		Children: make(map[string]*TreeNode),
	}

	var hexBuf [64]byte
	var childList []child

	for _, f := range node.files {
		hex.Encode(hexBuf[:], f.Sha256[:])
		buf := make([]byte, 0, 1+10+len(f.Path)+1+20+1+64+1)
		buf = append(buf, 0x00)
		buf = append(buf, "evtree:v1:"...)
		buf = append(buf, f.Path...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, f.Size, 10)
		buf = append(buf, ':')
		buf = append(buf, hexBuf[:]...)
		buf = append(buf, '\n')

		leafHash := Hash32(sha256.Sum256(buf))
		sha256Copy := f.Sha256
		result.Children[f.Path] = &TreeNode{
			Hash:   leafHash,
			Size:   f.Size,
			Sha256: &sha256Copy,
		}
		childList = append(childList, child{name: f.Path, hash: leafHash})
	}

	for name, sub := range node.subdirs {
		subNode := buildTreeNode(sub)
		result.Children[name] = subNode
		childList = append(childList, child{name: name, hash: subNode.Hash})
	}

	sort.Slice(childList, func(i, j int) bool {
		return childList[i].name < childList[j].name
	})

	if len(childList) == 0 {
		result.Hash = Hash32(sha256.Sum256([]byte{0x00}))
		return result
	}

	combined := make([]byte, 1, 1+32*len(childList))
	combined[0] = 0x01
	for _, c := range childList {
		combined = append(combined, c.hash[:]...)
	}
	result.Hash = Hash32(sha256.Sum256(combined))
	return result
}

func normalize(p string) string {
	p = filepath.ToSlash(p)
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}
