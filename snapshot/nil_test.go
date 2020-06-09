package snapshot

import (
	"testing"
	"unsafe"
)

func TestNil(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = NewNil()
	})
	if allocs != 0 {
		t.Errorf("NewNil should not alloc got: %f", allocs)
	}
	if unsafe.Sizeof(Nil{}) != 0 {
		t.Errorf("Nil should have size 0 got: %d", unsafe.Sizeof(Nil{}))
	}
}
