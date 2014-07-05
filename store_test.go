package memkv

import (
	"reflect"
	"testing"
)

var getTests = []struct {
	Key   string
	Value string
}{
	{"/myapp/database/username", "admin"},
	{"/myapp/database/password", "123456789"},
}

func TestGet(t *testing.T) {
	for _, tt := range getTests {
		s := New()
		s.Set(tt.Key, tt.Value)
		got, ok := s.Get(tt.Key)
		if !ok {
			t.Errorf("missing key")
		}
		want := Node{tt.Key, tt.Value}
		if got != want {
			t.Errorf("wanted %v, got %v", want, got)
		}
	}
}

type missingKeyResult struct {
	node Node
	ok   bool
}

func TestMissingKey(t *testing.T) {
	s := New()
	want := missingKeyResult{Node{}, false}
	node, ok := s.Get("/missing/key")
	got := missingKeyResult{node, ok}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

type globResult struct {
	nodes []Node
}

func TestGlob(t *testing.T) {
	s := New()
	s.Set("/myapp/database/username", "admin")
	s.Set("/myapp/database/password", "123456789")
	s.Set("/myapp/port", "80")
	want := globResult{
		nodes: []Node{
			Node{"/myapp/database/password", "123456789"},
			Node{"/myapp/database/username", "admin"},
		},
	}
	nodes, err := s.Glob("/myapp/*/*")
	if err != nil {
		t.Error(err.Error())
	}
	got := globResult{nodes}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

func TestGlobWithBadPattern(t *testing.T) {
	s := New()
	s.Set("/myapp/database/username", "admin")
	_, err := s.Glob("[]a]")
	if err == nil {
		t.Error("expected an error on bad pattern")
	}
}
