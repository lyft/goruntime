package snapshot

import (
	"encoding/binary"
	"hash/crc32"
	"math/rand"
	"sync"
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

func min(lhs uint64, rhs uint64) uint64 {
	if lhs < rhs {
		return lhs
	} else {
		return rhs
	}
}

// Random number generator. Implementations should be thread safe.
type RandomGenerator interface {
	// @return uint64 a new random number.
	Random() uint64
}

// Implementation of RandomGenerator that uses a time seeded random generator.
type randomGeneratorImpl struct {
	sync.Mutex
	random *rand.Rand
}

func (r *randomGeneratorImpl) Random() uint64 {
	r.Lock()
	v := uint64(r.random.Int63())
	r.Unlock()
	return v
}

var defaultRandomGenerator RandomGenerator = &randomGeneratorImpl{
	random: rand.New(rand.NewSource(time.Now().UnixNano())),
}

// Implementation of Snapshot for the filesystem loader.
type Snapshot struct {
	entries map[string]*entry.Entry
}

func New() (s *Snapshot) {
	s = &Snapshot{
		entries: make(map[string]*entry.Entry),
	}

	return
}

func (s *Snapshot) FeatureEnabled(key string, defaultValue uint64) bool {
	return defaultRandomGenerator.Random()%100 < min(s.GetInteger(key, defaultValue), 100)
}

// FeatureEnabledForID checks that the crc32 of the id and key's byte value falls within the mod of
// the 0-100 value for the given feature. Use this method for "sticky" features
func (s *Snapshot) FeatureEnabledForID(key string, id uint64, defaultPercentage uint32) bool {
	if e, ok := s.Entries()[key]; ok {
		if e.Uint64Valid {
			return enabled(id, uint32(e.Uint64Value), key)
		}
	}

	return enabled(id, defaultPercentage, key)
}

func (s *Snapshot) Get(key string) string {
	e, ok := s.entries[key]
	if ok {
		return e.StringValue
	} else {
		return ""
	}
}

func (s *Snapshot) GetInteger(key string, defaultValue uint64) uint64 {
	e, ok := s.entries[key]
	if ok && e.Uint64Valid {
		return e.Uint64Value
	} else {
		return defaultValue
	}
}

// GetModified returns the last modified timestamp for key. If key does not
// exist, the zero value for time.Time is returned.
func (s *Snapshot) GetModified(key string) time.Time {
	if e, ok := s.entries[key]; ok {
		return e.Modified
	}

	return time.Time{}
}

func (s *Snapshot) Keys() []string {
	ret := []string{}
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
	b = append(b, []byte(feature)...)

	return crc32.ChecksumIEEE(b)
}
