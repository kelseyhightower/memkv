package memkv

import (
	"path/filepath"
	"sync"
)

type Store struct {
	sync.RWMutex
	m map[string]Node
}

func New() Store {
	return Store{m: make(map[string]Node)}
}

func (s Store) Get(key string) (Node, bool) {
	s.RLock()
	n, ok := s.m[key]
	s.RUnlock()
	return n, ok
}

func (s Store) GetAll(pattern string) (Nodes, error) {
	ns := make(Nodes, 0)
	s.RLock()
	defer s.RUnlock()
	for _, n := range s.m {
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
	s.Lock()
	s.m[key] = Node{key, value}
	s.Unlock()
}
