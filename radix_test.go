package there

import (
	"fmt"
	"testing"
)

func TestRadix(t *testing.T) {
	r := New()

	err, ok := r.Insert("foo", nil)
	fail := ok && err != nil

	if fail {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	err, ok = r.Insert("foo/bar/baz", nil)
	if fail {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	err, ok = r.Insert("foo/baz/bar", nil)
	if fail {
		t.Fatalf(fmt.Sprintf("%v", err))
	}
}

func TestUpdateEdge(t *testing.T) {
	edges := []edge{
		{
			label: '/',
			node: &node{
				prefix: "/ba",
				edges: []edge{
					{
						label: 'r',
						node: &node{
							leaf: &leafNode{
								key: "foo/bar/baz",
								val: nil,
							},
							prefix: "r/baz",
						},
					},
					{
						label: 'z',
						node: &node{
							leaf: &leafNode{
								key: "foo/baz/bar",
								val: nil,
							},
							prefix: "z/bar",
						},
					},
				},
			},
		},
	}

	parent := &node{
		leaf: &leafNode{
			key: "foo",
			val: nil,
		},
		prefix: "foo",
		edges:  edges,
	}

	child := &node{
		prefix: "/b",
	}

	err := parent.updateEdge('/', child)
	if err != nil {
		t.Fatalf("fail")
	}
}

func TestGet(t *testing.T) {
	r := New()

	dummyHandler := func() {}
	r.Insert("/foo", dummyHandler)
	r.Insert("/foo/bar/baz", dummyHandler)
	r.Insert("/foo/baz/bar", dummyHandler)

	val, _ := r.Get("/foo")
	if val == nil {
		t.Fatalf("fail")
	}

	val, _ = r.Get("/foo/bar/baz")
	if val == nil {
		t.Fatalf("fail")
	}

	val, _ = r.Get("/foo/baz/bar")
	if val == nil {
		t.Fatalf("fail")
	}
}
