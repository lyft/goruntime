package snapshot

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
	"github.com/stretchr/testify/assert"
)

func TestRandomGeneratorImpl_Random_Race(t *testing.T) {
	snap := New()
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				snap.rand.Uint64()
			}
		}()
	}
	wg.Wait()
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

func BenchmarkCRC(b *testing.B) {
	const a = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const key = a + a
	if len(key) != 64 {
		panic(len(key))
	}
	b.SetBytes(int64(8 + len(key)))
	for i := 0; i < b.N; i++ {
		crc(uint64(i), key)
	}
}

func BenchmarkEnabled(b *testing.B) {
	const a = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const key = a + a
	if len(key) != 64 {
		panic(len(key))
	}
	b.SetBytes(int64(8 + len(key)))
	for i := 0; i < b.N; i++ {
		enabled(uint64(1), 50, key)
	}
}

func setupFeatureEnabled(b *testing.B, key string) *Snapshot {
	snap := New()
	snap.entries[key] = &entry.Entry{
		Uint64Value: 50,
		Uint64Valid: true,
	}
	b.ResetTimer()
	return snap
}

func BenchmarkFeatureEnabled(b *testing.B) {
	const key = "this_is_a_test_key"
	snap := setupFeatureEnabled(b, key)
	for i := 0; i < b.N; i++ {
		snap.FeatureEnabled(key, 100)
	}
}

func BenchmarkFeatureEnabled_Parallel(b *testing.B) {
	const key = "this_is_a_test_key"
	var snaps [4]*Snapshot
	for i := 0; i < 4; i++ {
		snaps[i] = setupFeatureEnabled(b, key)
	}
	n := new(int32)
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)
	b.RunParallel(func(pb *testing.PB) {
		i := int(atomic.AddInt32(n, 1))
		snap := snaps[i%len(snaps)]
		for pb.Next() {
			snap.FeatureEnabled(key, 100)
		}
	})
}
