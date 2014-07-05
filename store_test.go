package memkv

import (
	"path/filepath"
	"reflect"
	"testing"
)

var getTests = []struct {
	key   string
	value string
	ok    bool
	want  Node
}{
	{
		"/myapp/database/username",
		"admin",
		true,
		Node{"/myapp/database/username", "admin"},
	},
	{
		"/myapp/database/password",
		"123456789",
		true,
		Node{"/myapp/database/password", "123456789"},
	},
	{
		"/missing",
		"",
		false,
		Node{},
	},
}

func TestGet(t *testing.T) {
	for _, tt := range getTests {
		s := New()
		if tt.ok {
			s.Set(tt.key, tt.value)
		}
		got, ok := s.Get(tt.key)
		if ok != tt.ok {
			t.Errorf("wanted %v, got %v", tt.ok, got)
		}
		if got != tt.want {
			t.Errorf("wanted %v, got %v", tt.want, got)
		}
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
			"/myapp/port":                 "443",
			"/myapp/url":                  "app.example.com",
			"/myapp/upstream/app1":        "203.0.113.0.1:8080",
			"/myapp/upstream/app2":        "203.0.113.0.2:8080",
			"/myapp/upstream/app1/domain": "app.example.com",
			"/myapp/upstream/app2/domain": "app.example.com",
		},
		"/myapp/upstream/*",
		[]Node{
			Node{"/myapp/upstream/app1", "203.0.113.0.1:8080"},
			Node{"/myapp/upstream/app2", "203.0.113.0.2:8080"},
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
