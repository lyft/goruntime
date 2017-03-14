package snapshot

import (
	"math/rand"
	"testing"
	"time"
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
