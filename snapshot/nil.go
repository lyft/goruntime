package snapshot

import "github.com/lyft/goruntime/snapshot/entry"

// Implementation of Snapshot for the nilLoaderImpl.
type Nil struct{}

func NewNil() (s *Nil) {
	s = &Nil{}

	return
}

func (n *Nil) FeatureEnabled(key string, defaultValue uint64) bool {
	return defaultRandomGenerator.Random()%100 < min(defaultValue, 100)
}

func (n *Nil) Get(key string) string { return "" }

func (n *Nil) GetInteger(key string, defaultValue uint64) uint64 {
	return defaultValue
}

func (n *Nil) Keys() []string {
	return []string{}
}

func (n *Nil) Entries() map[string]*entry.Entry {
	return make(map[string]*entry.Entry)
}

func (n *Nil) SetEntry(string, *entry.Entry) {}
