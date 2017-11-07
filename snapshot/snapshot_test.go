package snapshot

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
	"github.com/stretchr/testify/assert"
)

func TestRandomGeneratorImpl_Random_Race(t *testing.T) {
	rgi := &randomGeneratorImpl{random: rand.New(rand.NewSource(time.Now().UnixNano()))}

	go func() {
		for i := 0; i < 100; i++ {
			rgi.Random()
		}
	}()

	for i := 0; i < 100; i++ {
		rgi.Random()
	}
}

func TestSnapshot_FeatureEnabledForID(t *testing.T) {
	key := "test"
	ss := NewMock()
	ss.SetUInt64(key, 100)
	assert.True(t, ss.FeatureEnabledForID(key, 1, 100))

	ss.SetUInt64(key, 0)
	assert.False(t, ss.FeatureEnabledForID(key, 1, 100))

	enabled := 0
	for i := 1; i < 101; i++ {
		ss.SetUInt64(key, uint64(i))
		if ss.FeatureEnabledForID(key, uint64(i), 100) {
			enabled++
		}
	}

	assert.Equal(t, 47, enabled)
}

func TestSnapshot_FeatureEnabledForIDDisabled(t *testing.T) {
	key := "test"
	ss := NewMock()
	assert.True(t, ss.FeatureEnabledForID(key, 1, 100))
	assert.False(t, ss.FeatureEnabledForID(key, 1, 0))
}

func TestSnapshot_GetModified(t *testing.T) {
	ss := NewMock()

	assert.True(t, ss.GetModified("foo").IsZero())

	now := time.Now()
	ss.entries["foo"] = &entry.Entry{Modified: now}
	assert.Equal(t, now, ss.GetModified("foo"))
}
