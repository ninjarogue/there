package there

import (
	"errors"
	"fmt"
	"log"
	"sort"
)



type leafNode struct {
	key string
	val any
}

type edge struct {
	label byte
	node *node
}

type edges []edge

func (e edges) Len() int {
	return len(e)
}

func (e edges) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges) Sort() {
	sort.Sort(e)
}

type node struct {
	leaf *leafNode
	prefix string
	edges edges
}

// Return true if the node is a "leaf node".
func (n *node) isLeaf() bool {
	return n.leaf != nil
}

func (n *node) addEdge(e edge) {
	n.edges = append(n.edges, e)
	n.edges.Sort()
}

// Returns nil if the edge was not found or else findEdge will return the node corresponding to the edge.
func (n *node) findEdge(label byte) *node {
	num := len(n.edges)

	// Find the index of the matching edge label.
	// If the value of "idx" equals the length of "num", a matching label was not found.
	idx := 0
	for idx < num {
		if n.edges[idx].label >= label {
			break
		}
		idx++
	}

	if idx < num && n.edges[idx].label == label {
		return n.edges[idx].node
	}

	return nil
}

func (n *node) updateEdge(label byte, node *node) error {
	num := len(n.edges)

	idx := 0
	for idx < num {
		if n.edges[idx].label >= label {
			break
		}
		idx++
	}

	if idx < num && n.edges[idx].label == label {
		n.edges[idx].node = node
		return nil
	}

	return errors.New("replacing missing edge")
}

type Tree struct {
	method string
	root *node
	size int
}

func New() *Tree {
	return &Tree{root: &node{}}
}

func (t *Tree) GET(s string, v any) (*Tree, error) {
	t.method = "GET"
	// Insert key and return method tree.
	_, err := t.Insert(s, v)

	if err != nil {
		return t, err
	}

	return t, nil
}

func (t *Tree) POST(s string, v any) (*Tree, error) {
	t.method = "POST"
	// Insert key and return method tree.
	_, err := t.Insert(s, v)

	if err != nil {
		return t, err
	}

	return t, nil
}

func (t *Tree) Insert(s string, v any) (any, error) {
	var parent *node
	n := t.root
	search := s

	for {
		// Handle key exhaustion.
		// This code block also allows us to deal with duplicate keys.
		if len(search) == 0 {
			if n.isLeaf() {
				old := n.leaf.val
				n.leaf.val = v
				return old, nil
			}

			n.leaf = &leafNode{
				key: s,
				val: v,
			}

			t.size++
			return nil, nil
		}

		// Look for the edge.
		parent = n
		n = n.findEdge(search[0])

		// No edge? Create one.
		if n == nil {
			e := edge {
				label: search[0],
				node: &node{
					leaf: &leafNode{
						key: s,
						val: v,
					},
					prefix: search,
				},
			}

			parent.addEdge(e)
			t.size++
			return nil, nil
		}

		// Find longest prefix of the search "key" on match.
		commonPrefix := longestPrefix(search, n.prefix)
		// If the prefix matches the commonPrefix we continue traversing the tree.
		// We reassign "search" to the remaining portion of the last search string
		// to continue searching for a place to insert.
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:] // NOTE: This statement could possibly exhaust the search key
			continue
		}

		// Split the node.
		t.size++
		child := &node {
			prefix: search[:commonPrefix],
		}
		err := parent.updateEdge(search[0], child)

		if err != nil {
			return nil, err
		}

		// Restore the existing node.
		child.addEdge(edge {
			label: n.prefix[commonPrefix],
			node: n,
		})
		n.prefix = n.prefix[commonPrefix:]

		// Create a new leaf node.
		leaf := &leafNode{
			key: s,
			val: v,
		}

		// If the new key is a subset, add to to this node.
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.leaf = leaf
			return nil, nil
		}

		// Create a new edge for the node.
		child.addEdge(edge{
			label: search[0],
			node: &node{
				leaf: leaf,
				prefix: search,
			},
		})

		return nil, nil
	}
}

func longestPrefix(k1, k2 string) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}

	var idx int
	for idx = 0; idx < max; idx++ {
		if k1[idx] != k2[idx] {
			break
		}
	}

	return idx
}
