package snapshot

import (
	"math/rand"
	"testing"
	"time"

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
	assert.True(t, ss.FeatureEnabledForID(key, 1))

	ss.SetUInt64(key, 0)
	assert.False(t, ss.FeatureEnabledForID(key, 1))

	enabled := 0
	for i := 1; i < 101; i++ {
		ss.SetUInt64(key, uint64(i))
		if ss.FeatureEnabledForID(key, uint64(i)) {
			enabled++
		}
	}

	assert.Equal(t, 47, enabled)
}
