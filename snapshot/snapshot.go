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
func (s *Snapshot) FeatureEnabledForID(key string, id uint64, defaultValue bool) bool {
	if e, ok := s.Entries()[key]; ok {
		if e.Uint64Valid {
			return withinPercentile(id, uint32(e.Uint64Value), key)
		}
	}

	return defaultValue
}

func (s *Snapshot) Get(key string) string {
	entry, ok := s.entries[key]
	if ok {
		return entry.StringValue
	} else {
		return ""
	}
}

func (s *Snapshot) GetInteger(key string, defaultValue uint64) uint64 {
	entry, ok := s.entries[key]
	if ok && entry.Uint64Valid {
		return entry.Uint64Value
	} else {
		return defaultValue
	}
}

func (s *Snapshot) Keys() []string {
	ret := []string{}
	for key, _ := range s.entries {
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

func withinPercentile(id uint64, percentage uint32, feature string) bool {
	uid := crc(id, feature)

	m := uid % 100

	return m < percentage
}

func crc(id uint64, feature string) uint32 {
	b := make([]byte, 8, len(feature)+8)
	binary.LittleEndian.PutUint64(b, id)
	b = append(b, []byte(feature)...)

	return crc32.ChecksumIEEE(b)
}
