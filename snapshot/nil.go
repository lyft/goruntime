package snapshot

import (
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

// Implementation of Snapshot for the nilLoaderImpl.
type Nil struct{}

func NewNil() Nil { return Nil{} }

func (Nil) FeatureEnabled(_ string, defaultValue uint64) bool {
	return defaultRandomGenerator.Random()%100 < min(defaultValue, 100)
}

func (Nil) FeatureEnabledForID(string, uint64, uint32) bool {
	return true
}

func (Nil) Get(string) string {
	return ""
}

func (Nil) GetInteger(_ string, defaultValue uint64) uint64 {
	return defaultValue
}

func (Nil) GetModified(string) time.Time {
	return time.Time{}
}

func (Nil) Keys() []string {
	return []string{}
}

func (Nil) Entries() map[string]*entry.Entry {
	return map[string]*entry.Entry{}
}

func (Nil) SetEntry(string, *entry.Entry) {}
