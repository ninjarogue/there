package there

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type route struct {
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
	leaf *route
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

type methodTree struct {
	method string
	root *node
	size int
}

type Base struct {
	methodTrees map[string]methodTree
}

func (b *Base) AddRoute(r *Route) (*Base, error) {
	// Return an error if a route method is not a valid http method.
	for _, m := range r.Methods {
		m = strings.ToUpper(m)
		if v, found := httpMethods[m]; !found {
			return b, errors.New(fmt.Sprintf("invalid http method: %v", v))
		}

		// If the method tree does not already exist create it.
		if _, found := b.methodTrees[m]; !found {
			mt := &methodTree{
				method: m,
				root: &node{
					prefix: "",
				},
				size: 0,
			}

			mt.Insert(r) // TODO: Handle error.
			b.methodTrees[m] = *mt
			continue
		}

		// If the method tree already exists add the new entry to the tree.
		t, _ := b.methodTrees[m]
		t.Insert(r)
	}

	return b, nil
}

// TODO: Add more http methods.
var httpMethods = map[string]string{
	"GET": "GET",
	"POST": "POST",
	"DELETE": "DELETE",
	"PUT": "PUT",
	"PATCH": "PATCH",
}

func (t *methodTree) Insert(r *Route) (any, error) {
	var parent *node
	n := t.root

	s := r.stringPath
	search := s

	for {
		// Handle key exhaustion.
		// This code block also allows us to deal with duplicate keys.
		if len(search) == 0 {
			if n.isLeaf() {
				old := n.leaf.endpoint
				n.leaf.endpoint = r.Endpoint
				return old, nil
			}

			n.leaf = &route{
				key: s,
				endpoint: r.Endpoint,
			}

			t.size++
			return nil, nil
		}

		// Look for the edge.
		parent = n
		n = n.findEdge(search[0])

		// No edge? Create one.
		if n == nil {
			// Trim the trailing slash (if any), unless the search key is equal to "/".
			key := s
			if s != "/" {
				key = strings.TrimSuffix(s, "/")
			}

			e := edge {
				label: search[0],
				node: &node{
					leaf: &route{
						// FIX: Do we need to include the key field in the route struct?
						key: key,
						Path: r.Path,
						middlewares: r.Middlewares,
						endpoint: r.Endpoint,
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
		leaf := &route{
			key: s,
			Path: r.Path,
			middlewares: r.Middlewares,
			endpoint: r.Endpoint,
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
func (b *Base) LookUp(method string, s string) (*route, map[string]string, bool) {
	t := b.methodTrees[method]
	n := t.root
	search := s
	var routeParams map[string]string

	for {
		// Handle key exhaustion.
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf, routeParams, true
			}
			break
		}

		// Look for an edge
		n = n.findEdge(search[0])
		if n == nil {
			break
		}

		// Check to see if the current route segment is a variable.
		if n.variable {
			params, ok := n.leaf.Path.Parse(s)

			if routeParams != nil && ok {
				routeParams = params
				for _, v := range routeParams {
					if search[:len(v)] == v {
						search = search[:len(v)]
					}
				}
				continue
			}
		}

		// If we find a match, we truncate
		// the matching slice and continue the search.
		if strings.HasPrefix(search, n.prefix) {
			// Consume the current route segment.
			search = search[len(n.prefix):]
		} else {
			break
		}
	}

	return n.leaf, routeParams, false
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
