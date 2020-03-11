package snapshot

import (
	"math/rand"
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

// Implementation of Snapshot for the nilLoaderImpl.
type Nil struct{}

func NewNil() (s *Nil) {
	return &Nil{}
}

var nilRandom = random{
	rr: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func (n *Nil) FeatureEnabled(key string, defaultValue uint64) bool {
	return nilRandom.Uint64()%100 < min(defaultValue, 100)
}

func (n *Nil) FeatureEnabledForID(key string, id uint64, defaultPercentage uint32) bool {
	return true
}

func (n *Nil) Get(key string) string { return "" }

func (n *Nil) GetInteger(key string, defaultValue uint64) uint64 { return defaultValue }

func (n *Nil) GetModified(key string) time.Time { return time.Time{} }

func (n *Nil) Keys() []string {
	return []string{}
}

func (n *Nil) Entries() map[string]*entry.Entry {
	return make(map[string]*entry.Entry)
}

func (n *Nil) SetEntry(string, *entry.Entry) {}
