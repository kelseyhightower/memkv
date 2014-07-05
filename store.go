package memkv

import (
	"path/filepath"
)

type Store map[string]Node

func New() Store {
	return make(Store)
}

func (s Store) Get(key string) (Node, bool) {
	n, ok := s[key]
	return n, ok
}

func (s Store) GetAll(pattern string) (Nodes, error) {
	ns := make(Nodes, 0)
	for _, n := range s {
		m, err := filepath.Match(pattern, n.Key)
		if err != nil {
			return nil, err
		}
		if m {
			ns = append(ns, n)
		}
	}
	return ns, nil
}

func (s Store) Set(key string, value string) {
	s[key] = Node{key, value}
}
