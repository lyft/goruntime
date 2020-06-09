package loader

import "github.com/lyft/goruntime/snapshot"

// Implementation of Loader with no backing store.
type Nil struct {
	snapshot snapshot.Nil
}

func NewNil() Nil {
	return Nil{}
}

func (n Nil) Snapshot() snapshot.IFace { return n.snapshot }

func (Nil) AddUpdateCallback(callback chan<- int) {}
