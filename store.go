// Copyright 2014 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by a BSD-style
// license found in the LICENSE file.

// Package memkv implements an in-memory key/value store.
package memkv

import (
	"errors"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var ErrNotExist = errors.New("key does not exist")
var ErrNoMatch = errors.New("no keys match")

// A Store represents an in-memory key-value store safe for
// concurrent access.
type Store struct {
	FuncMap map[string]interface{}
	sync.RWMutex
	m map[string]KVPair
}

// New creates and initializes a new Store.
func New() Store {
	s := Store{m: make(map[string]KVPair)}
	s.FuncMap = map[string]interface{}{
		"exists": s.Exists,
		"ls":     s.List,
		"lsdir":  s.ListDir,
		"get":    s.Get,
		"gets":   s.GetAll,
		"getv":   s.GetValue,
		"getvs":  s.GetAllValues,
	}
	return s
}

// Delete deletes the KVPair associated with key.
func (s Store) Del(key string) {
	s.Lock()
	delete(s.m, key)
	s.Unlock()
}

// Exists checks for the existence of key in the store.
func (s Store) Exists(key string) bool {
	kv := s.Get(key)
	if kv.Value == "" {
		return false
	}
	return true
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}.
func (s Store) Get(key string) KVPair {
	s.RLock()
	kv := s.m[key]
	s.RUnlock()
	return kv
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in filepath.Match.
func (s Store) GetAll(pattern string) KVPairs {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		m, err := filepath.Match(pattern, kv.Key)
		if err != nil {
			return nil
		}
		if m {
			ks = append(ks, kv)
		}
	}
	if len(ks) == 0 {
		return nil
	}
	sort.Sort(ks)
	return ks
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "".
func (s Store) GetValue(key string) string {
	return s.Get(key).Value
}

func (s Store) GetAllValues(pattern string) []string {
	vs := make([]string, 0)
	for _, kv := range s.GetAll(pattern) {
		vs = append(vs, kv.Value)
	}
	sort.Strings(vs)
	return vs
}

func (s Store) List(filePath string) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		if kv.Key == filePath {
			m[path.Base(kv.Key)] = true
			continue
		}
		if strings.HasPrefix(kv.Key, filePath) {
			m[strings.Split(stripKey(kv.Key, filePath), "/")[0]] = true
		}
	}
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

func (s Store) ListDir(filePath string) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		if strings.HasPrefix(kv.Key, filePath) {
			items := strings.Split(stripKey(kv.Key, filePath), "/")
			if len(items) < 2 {
				continue
			}
			m[items[0]] = true
		}
	}
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

// Set sets the KVPair entry associated with key to value.
func (s Store) Set(key string, value string) {
	s.Lock()
	s.m[key] = KVPair{key, value}
	s.Unlock()
}

func (s Store) Purge() {
	s.Lock()
	for k := range s.m {
		delete(s.m, k)
	}
	s.Unlock()
}

func stripKey(key, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(key, prefix), "/")
}
