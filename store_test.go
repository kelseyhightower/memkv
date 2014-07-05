package memkv

import (
	"path/filepath"
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

var globTests = []struct {
	input   map[string]string
	pattern string
	want    []Node
	err     error
}{
	{
		map[string]string{
			"/myapp/database/password": "123456789",
			"/myapp/database/username": "admin",
		},
		"/myapp/*/*",
		[]Node{
			Node{"/myapp/database/password", "123456789"},
			Node{"/myapp/database/username", "admin"},
		},
		nil,
	},
	{
		map[string]string{
			"/myapp/database/password": "123456789",
			"/myapp/database/username": "admin",
		},
		"[]a]",
		nil,
		filepath.ErrBadPattern,
	},
}

func TestGlob(t *testing.T) {
	for _, tt := range globTests {
		s := New()
		for k, v := range tt.input {
			s.Set(k, v)
		}
		want := globResult{tt.want}
		nodes, err := s.Glob(tt.pattern)
		if err != tt.err {
			t.Error(err.Error())
		}
		got := globResult{nodes}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("wanted %v, got %v", want, got)
		}
	}
}
