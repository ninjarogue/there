package there

import (
	"errors"
	"log"
	"sort"
	"strings"
)

type leafNode struct {
	key string
	Path Path
	middlewares []Middleware
	endpoint Endpoint
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
	variable bool
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

type MethodTree struct {
	method string
	root *node
	size int
}

func New() *MethodTree {
	return &MethodTree{root: &node{}}
}

func (t *MethodTree) GET(s string, ep Endpoint) (*MethodTree, error) {
	t.method = "GET"

	_, err := t.Insert(s, ep)

	if err != nil {
		return t, err
	}

	return t, nil
}

func (t *MethodTree) POST(s string, ep Endpoint) (*MethodTree, error) {
	t.method = "POST"

	_, err := t.Insert(s, ep)

	if err != nil {
		return t, err
	}

	return t, nil
}

func (t *MethodTree) Insert(s string, ep Endpoint) (any, error) {
	var parent *node
	n := t.root
	search := s

	for {
		// Handle key exhaustion.
		// This code block also allows us to deal with duplicate keys.
		if len(search) == 0 {
			if n.isLeaf() {
				old := n.leaf.endpoint
				n.leaf.endpoint = ep
				return old, nil
			}

			n.leaf = &leafNode{
				key: s,
				endpoint: ep,
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
						endpoint: ep,
					},
					prefix: search,
				},
			}

			parent.addEdge(e)
			t.size++
			return nil, nil
		}

		// Find longest prefix of the search "key" on match.
		commonPrefix := longestCommonPrefix(search, n.prefix)
		// If the prefix matches the commonPrefix we continue traversing the tree.
		// We reassign "search" to the remaining portion of the last search string
		// to continue searching for a place to insert.
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// Split the node.
		t.size++
		child := &node {
			prefix: search[:commonPrefix],
		}
		if strings.HasPrefix(search[:commonPrefix], ":") {
			child.variable = true
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
			endpoint: ep,
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

// Get is used to lookup a specific key
// returning the value if it was found.
func (t *MethodTree) Get(s string) (Endpoint, bool) {
	n := t.root
	search := s

	for {
		// Handle key exhaustion.
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf.endpoint, true
			}
			break
		}

		// Look for an edge
		n = n.findEdge(search[0])
		if n == nil {
			break
		}

		// Check to see if the current route segment is a variable
		if n.variable {
			log.Printf("%v", n.leaf.key)
			// TODO: ...
		}

		// If we find a match, we truncate
		// the matching slice and continue the search.
		if strings.HasPrefix(search, n.prefix) {
			// Consume the search key.
			search = search[len(n.prefix):]
		} else {
			break
		}
	}

	return nil, false
}

// TODO: Use the original Parse method.
func Parse(p Path, route string) (map[string]string, bool) {
	params := map[string]string{}

	split := splitUrl(route)

	if len(split) != len(p.parts) {
		return nil, false
	}

	ignoreCase := p.ignoreCase

	for i := 0; i < len(p.parts); i++ {
		a := p.parts[i]
		b := split[i]
		if a.variable {
			params[a.value] = b
		} else {
			if (ignoreCase && strings.ToLower(a.value) != strings.ToLower(b)) ||
				(!ignoreCase && a.value != b) {
				return nil, false
			}
		}
	}

	return params, true
}

func longestCommonPrefix(k1, k2 string) int {
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
