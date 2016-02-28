// Copyright 2014 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by a BSD-style
// license found in the LICENSE file.

// Package memkv implements an in-memory key/value store.
package memkv

import (
	"encoding/json"
	"errors"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var ErrNotExist = errors.New("key does not exist")
var ErrNoMatch = errors.New("no keys match")

type KeyError struct {
	Key string
	Err error
}

func (e *KeyError) Error() string {
	return e.Err.Error() + ": " + e.Key
}

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
		"getrx":  s.GetAllRegexp,
		"group":  s.Group,
		"select": s.Select,
		"hash":   s.Hash,
		"take":   s.Take,
		"keys":   s.Keys,
		"values": s.Values,
		"object": s.JSONObject,
		"array":  s.JSONArray,
		"join":   s.Join,
	}
	return s
}

// Delete deletes the KVPair associated with key.
func (s Store) Del(key string) {
	s.DelAll(key)
}

func (s Store) DelAll(pattern string) error {
	ks, err := s.GetAll(pattern)
	s.Lock()
	defer s.Unlock()
	if err != nil {
		return err
	}
	for _, kv := range ks {
		delete(s.m, kv.Key)
	}
	return nil
}

func (s Store) DelAllRegexp(pattern string) error {
	ks, err := s.GetAllRegexp(pattern)
	s.Lock()
	defer s.Unlock()
	if err != nil {
		return err
	}
	for _, kv := range ks {
		delete(s.m, kv.Key)
	}
	return nil
}

// Exists checks for the existence of key in the store.
func (s Store) Exists(key string) bool {
	_, err := s.Get(key)
	if err != nil {
		return false
	}
	return true
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}, ErrNotExist.
func (s Store) Get(key string) (KVPair, error) {
	s.RLock()
	kv, ok := s.m[key]
	s.RUnlock()
	if !ok {
		return kv, &KeyError{key, ErrNotExist}
	}
	return kv, nil
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "", ErrNotExist.
func (s Store) GetValue(key string, v ...string) (string, error) {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	kv, err := s.Get(key)
	if err != nil {
		if defaultValue != "" {
			return defaultValue, nil
		}
		return "", err
	}
	return kv.Value, nil
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in filepath.Match.
func (s Store) GetAll(pattern string) (KVPairs, error) {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		m, err := filepath.Match(pattern, kv.Key)
		if err != nil {
			return nil, err
		}
		if m {
			ks = append(ks, kv)
		}
	}
	if len(ks) == 0 {
		return ks, nil
	}
	sort.Sort(ks)
	return ks, nil
}

func (s Store) GetAllValues(pattern string) ([]string, error) {
	vs := make([]string, 0)
	ks, err := s.GetAll(pattern)
	if err != nil {
		return vs, err
	}
	if len(ks) == 0 {
		return vs, nil
	}
	for _, kv := range ks {
		vs = append(vs, kv.Value)
	}
	sort.Strings(vs)
	return vs, nil
}

func (s Store) List(filePath string) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	s.RLock()
	defer s.RUnlock()
	prefix := pathToTerms(path.Clean(filePath))
	for _, kv := range s.m {
		if kv.Key == filePath {
			m[path.Base(kv.Key)] = true
			continue
		}
		target := pathToTerms(path.Dir(kv.Key))
		if samePrefixTerms(target, prefix) {
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
	prefix := pathToTerms(path.Clean(filePath))
	for _, kv := range s.m {
		if strings.HasPrefix(kv.Key, filePath) {
			items := pathToTerms(path.Dir(kv.Key))
			if samePrefixTerms(prefix, items) && (len(items)-len(prefix) >= 1) {
				m[items[len(prefix):][0]] = true
			}
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

func (s Store) GetAllRegexp(pattern string) (KVPairs, error) {
	ks := make(KVPairs, 0)
	re := regexp.MustCompile(pattern)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		if re.MatchString(kv.Key) {
			ks = append(ks, kv)
		}
	}
	if len(ks) == 0 {
		return ks, nil
	}
	sort.Sort(ks)
	return ks, nil
}

func (s Store) Group(ks KVPairs, pattern string) map[string]KVPairs {
	re := regexp.MustCompile(pattern)
	result := make(map[string]KVPairs)
	for _, kv := range ks {
		match := re.FindStringSubmatch(kv.Key)
		if len(match) >= 1 {
			result[match[1]] = append(result[match[1]], kv)
		}
	}
	for group, _ := range result {
		sort.Sort(result[group])
	}
	return result
}

func (s Store) Hash(ks KVPairs, pattern string) map[string]string {
	re := regexp.MustCompile(pattern)
	result := make(map[string]string)
	for _, kv := range ks {
		match := re.FindStringSubmatch(kv.Key)
		if len(match) >= 1 {
			result[match[1]] = kv.Value
		}
	}
	return result
}

func (s Store) Select(ks KVPairs, pattern string) KVPairs {
	re := regexp.MustCompile(pattern)
	result := KVPairs{}
	for _, kv := range ks {
		if re.MatchString(kv.Key) {
			result = append(result, kv)
		}
	}
	sort.Sort(result)
	return result
}

func (s Store) Take(ks KVPairs, pattern string) string {
	re := regexp.MustCompile(pattern)
	for _, kv := range ks {
		if re.MatchString(kv.Key) {
			return kv.Value
		}
	}
	return ""
}

func (s Store) Keys(ks KVPairs) []string {
	result := make([]string, 0)
	for _, kv := range ks {
		result = append(result, kv.Key)
	}
	return result
}

func (s Store) Values(ks KVPairs) []string {
	result := make([]string, 0)
	for _, kv := range ks {
		result = append(result, kv.Value)
	}
	return result
}

func (s Store) JSONObject(raw string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(raw), &data)
	return data, err
}

func (s Store) JSONArray(raw string) ([]interface{}, error) {
	var data []interface{}
	err := json.Unmarshal([]byte(raw), &data)
	return data, err
}

func (s Store) Join(strs ...string) string {
	return strings.Join(strs, "")
}

func stripKey(key, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(key, prefix), "/")
}

func pathToTerms(filePath string) []string {
	return strings.Split(path.Clean(filePath), "/")
}

func samePrefixTerms(left, right []string) bool {
	l := len(left)
	if len(left) > len(right) {
		l = len(right)
	}
	for i := 0; i < l; i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
