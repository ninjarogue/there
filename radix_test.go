package there

import (
	"fmt"
	"testing"
)

func TestRadix(t *testing.T) {
	r := New()

	_, err := r.Insert("foo", nil)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	_, err = r.Insert("foo/bar/baz", nil)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	_, err = r.Insert("foo/baz/bar", nil)
	if err != nil {
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
								Path: Path{
									parts: []pathPart{
										{value: "foo", variable: false,},
										{value: "bar", variable: false,},
										{value: "baz", variable: false,},
									},
									ignoreCase: false,
								},
							},
							prefix: "r/baz",
						},
					},
					{
						label: 'z',
						node: &node{
							leaf: &leafNode{
								Path: Path{
									parts: []pathPart{
										{value: "foo", variable: false,},
										{value: "baz", variable: false,},
										{value: "bar", variable: false,},
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

	parent := &node{
		leaf: &leafNode{
			Path: Path{
				parts: []pathPart{
					{value: "foo", variable: false,},
				},
				ignoreCase: false,
			},
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

	dummyHandler := func(r Request) Response { return nil }
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
