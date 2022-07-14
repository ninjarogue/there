package there

import (
	"errors"
	"sort"
	"strings"
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
	root *node
	size int
}

func New() *Tree {
	return &Tree{root: &node{}}
}

func (t *Tree) Insert(s string, v any) (any, bool) {
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
				return old, true
			}

			n.leaf = &leafNode{
				key: s,
				val: v,
			}

			t.size++
			return nil, false
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
			return nil, false
		}

		// Find longest prefix of the search "key" on match.
		commonPrefix := longestPrefix(search, n.prefix)
		// If the prefix matches the commonPrefix we continue traversing the tree.
		// We reassign "search" to the remaining portion of the last search string
		// to continue searching for a place to insert.
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:] // We exhaust/consume the search key.
			continue
		}

		// Split the node.
		t.size++
		child := &node {
			prefix: search[:commonPrefix],
		}
		err := parent.updateEdge(search[0], child)

		if err != nil {
			return err, false
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
			return nil, false
		}

		// Create a new edge for the node.
		child.addEdge(edge{
			label: search[0],
			node: &node{
				leaf: leaf,
				prefix: search,
			},
		})

		return nil, false
	}
}

// Get is used to lookup a specific key
// returning the value if it was found.
func (t *Tree) Get(s string) (any, bool) {
	n := t.root
	search := s

	for {
		// Handle key exhaustion.
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf.val, true
			}
			break
		}

		// Look for an edge
		n = n.findEdge(search[0])
		if n == nil {
			break
		}

		// If we find a match, we truncate
		// the matching slice and continue the search.
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):] // We exhaust/consume the search key.
		} else {
			break
		}
	}

	return nil, false
}

func longestPrefix(k1, k2 string) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}

	idx := 0
	for idx < max {
		if k1[idx] != k2[idx] {
			break
		}
		idx++
	}

	return idx
}
