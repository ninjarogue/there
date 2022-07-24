package there

import (
	"testing"
)

func TestRadix(t *testing.T) {
	r := NewRouter()

	r.Get("foo", nil)
	r.Get("foo/bar/baz", nil)
	r.Get("foo/baz/bar", nil)

	// TODO: Test "r".
}

func TestUpdateEdge(t *testing.T) {
	edges := []edge{
		{
			label: '/',
			node: &Node{
				prefix: "/ba",
				edges: []edge{
					{
						label: 'r',
						node: &Node{
							leaf: &LeafNode{
								Path: Path{
									parts: []pathPart{
										{value: "foo", variable: false},
										{value: "bar", variable: false},
										{value: "baz", variable: false},
									},
									ignoreCase: false,
								},
							},
							prefix: "r/baz",
						},
					},
					{
						label: 'z',
						node: &Node{
							leaf: &LeafNode{
								Path: Path{
									parts: []pathPart{
										{value: "foo", variable: false},
										{value: "baz", variable: false},
										{value: "bar", variable: false},
									},
									ignoreCase: false,
								},
							},
							prefix: "z/bar",
						},
					},
				},
			},
		},
	}

	parent := &Node{
		leaf: &LeafNode{
			Path: Path{
				parts: []pathPart{
					{value: "foo", variable: false},
				},
				ignoreCase: false,
			},
		},
		prefix: "foo",
		edges:  edges,
	}

	child := &Node{
		prefix: "/b",
	}

	err := parent.updateEdge('/', child)
	if err != nil {
		t.Fatalf("fail")
	}
}

func TestGet(t *testing.T) {
	r := NewRouter()

	dummyHandler := func(r Request) Response { return nil }
	r.Get("/foo", dummyHandler)
	r.Get("/foo/bar/baz", dummyHandler)
	r.Get("/foo/baz/bar", dummyHandler)

	val, _, _ := r.base.LookUp("GET", "/foo")
	if val == nil {
		t.Fatalf("fail")
	}

	val, _, _ = r.base.LookUp("GET", "/foo/bar/baz")
	if val == nil {
		t.Fatalf("fail")
	}

	val, _, _ = r.base.LookUp("GET", "/foo/baz/bar")
	if val == nil {
		t.Fatalf("fail")
	}
}
