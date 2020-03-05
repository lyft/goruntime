package snapshot

import (
	"encoding/binary"
	"hash/crc32"
	"math/rand"
	"sync"
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

type random struct {
	mu sync.Mutex
	rr *rand.Rand
}

func (r *random) Uint64() uint64 {
	r.mu.Lock()
	x := r.rr.Int63()
	r.mu.Unlock()
	return uint64(x)
}

// Implementation of Snapshot for the filesystem loader.
type Snapshot struct {
	entries map[string]*entry.Entry
	rand    random
}

func New() (s *Snapshot) {
	rr := random{
		rr: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &Snapshot{
		entries: make(map[string]*entry.Entry),
		rand:    rr,
	}
}

func min(lhs, rhs uint64) uint64 {
	if lhs <= rhs {
		return lhs
	}
	return rhs
}

func (s *Snapshot) FeatureEnabled(key string, defaultValue uint64) bool {
	return s.rand.Uint64()%100 < min(s.GetInteger(key, defaultValue), 100)
}

// FeatureEnabledForID checks that the crc32 of the id and key's byte value falls within the mod of
// the 0-100 value for the given feature. Use this method for "sticky" features
func (s *Snapshot) FeatureEnabledForID(key string, id uint64, defaultPercentage uint32) bool {
	if e := s.entries[key]; e != nil && e.Uint64Valid {
		return enabled(id, uint32(e.Uint64Value), key)
	}
	return enabled(id, defaultPercentage, key)
}

func (s *Snapshot) Get(key string) string {
	if e := s.entries[key]; e != nil {
		return e.StringValue
	}
	return ""
}

func (s *Snapshot) GetInteger(key string, defaultValue uint64) uint64 {
	if e := s.entries[key]; e != nil && e.Uint64Valid {
		return e.Uint64Value
	}
	return defaultValue
}

// GetModified returns the last modified timestamp for key. If key does not
// exist, the zero value for time.Time is returned.
func (s *Snapshot) GetModified(key string) time.Time {
	if e := s.entries[key]; e != nil {
		return e.Modified
	}
	return time.Time{}
}

func (s *Snapshot) Keys() []string {
	ret := make([]string, 0, len(s.entries))
	for key := range s.entries {
		ret = append(ret, key)
	}
	return ret
}

func (s *Snapshot) Entries() map[string]*entry.Entry {
	return s.entries
}

func (s *Snapshot) SetEntry(key string, e *entry.Entry) {
	s.entries[key] = e
}

func enabled(id uint64, percentage uint32, feature string) bool {
	uid := crc(id, feature)
	return uid%100 < percentage
}

func crc(id uint64, feature string) uint32 {
	b := make([]byte, 8, len(feature)+8)
	binary.LittleEndian.PutUint64(b, id)
	return crc32.ChecksumIEEE(append(b, feature...))
}
