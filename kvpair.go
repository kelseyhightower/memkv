package memkv

import "strings"

type KVPair struct {
	Key   string
	Value string
}

type KVPairs []KVPair

func (ks KVPairs) Len() int {
	return len(ks)
}

func (ks KVPairs) Less(i, j int) bool {
	return ks[i].Key < ks[j].Key
}

func (ks KVPairs) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}

func (kv KVPair) Depth() int {
	return strings.Count(kv.Key, "/")
}

type DepthSortedKVPairs []KVPair

func (ks DepthSortedKVPairs) Len() int {
	return len(ks)
}

func (ks DepthSortedKVPairs) Less(i, j int) bool {
	if ks[i].Depth() == ks[j].Depth() {
		return ks[i].Key < ks[j].Key
	}
	return ks[i].Depth() < ks[j].Depth()
}

func (ks DepthSortedKVPairs) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}
