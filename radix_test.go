package there

import "testing"



func TestRadix(t *testing.T) {
	r := New()

	_, ok := r.Insert("foo", nil)
	if ok {
		t.Fatalf("fail")
	}

	_, ok = r.Insert("foo/bar/baz", nil)
	if ok {
		t.Fatalf("fail")
	}

	_, ok = r.Insert("foo/baz/bar", nil)
	if ok {
		t.Fatalf("fail")
	}
}
