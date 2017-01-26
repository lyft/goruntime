package loader

import "github.com/lyft/goruntime/snapshot"

// Implementation of Loader with no backing store.
type Nil struct {
	snapshot *snapshot.Nil
}

func NewNil() (n *Nil) {
	n = &Nil{
		snapshot: snapshot.NewNil(),
	}

	return
}

func (n *Nil) Snapshot() snapshot.IFace { return n.snapshot }

func (n *Nil) AddUpdateCallback(callback chan<- int) {}
