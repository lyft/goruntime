package loader

import (
	"fmt"

	"github.com/lyft/goruntime/snapshot"
	"github.com/lyft/goruntime/snapshot/entry"
	stats "github.com/lyft/gostats"
	"github.com/lyft/gostats/mock"
)

type MockOption interface {
	apply(*Loader)
}

type optionFunc func(*Loader)

func (f optionFunc) apply(ld *Loader) {
	f(ld)
}

func MockWithValues(kvs ...interface{}) MockOption {
	if len(kvs)&1 != 0 {
		panic(fmt.Sprintf("odd number of key values: %d: %v", len(kvs), kvs))
	}
	sn := snapshot.NewMock()
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i+0].(string)
		if !ok {
			panic(fmt.Sprintf("key %[1]v must be string got: %[1]T", key))
		}
		val := kvs[i+1]
		if s, ok := val.(string); ok {
			sn.Set(key, s)
			continue
		}
		var u uint64
		switch v := val.(type) {
		case int:
			u = uint64(v)
		case int8:
			u = uint64(v)
		case int16:
			u = uint64(v)
		case int32:
			u = uint64(v)
		case int64:
			u = uint64(v)
		case uint:
			u = uint64(v)
		case uint8:
			u = uint64(v)
		case uint16:
			u = uint64(v)
		case uint32:
			u = uint64(v)
		case uint64:
			u = uint64(v)
		default:
			panic(fmt.Sprintf("invalid type for key (%[1]s) value (%[2]v): %[2]T", key, val))
		}
		sn.SetUInt64(key, u)
	}
	return MockWithSnapshot(sn)
}

func MockWithEntries(ents map[string]*entry.Entry) MockOption {
	return optionFunc(func(ld *Loader) {
		ld.nextSnapshot = snapshot.New()
		for k, e := range ents {
			ld.nextSnapshot.SetEntry(k, e)
		}
	})
}

func MockWithSnapshot(sn *snapshot.Mock) MockOption {
	return optionFunc(func(ld *Loader) {
		ld.nextSnapshot = sn
	})
}

func MockWithScope(scope stats.Scope) MockOption {
	return optionFunc(func(ld *Loader) {
		ld.stats = newLoaderStats(scope)
	})
}

func MockWithSink(sink stats.Sink) MockOption {
	return MockWithScope(stats.NewStore(sink, false))
}

func MockWithUpdateChannel(changes <-chan *snapshot.Mock) MockOption {
	return optionFunc(func(ld *Loader) {
		if changes != nil {
			go func() {
				for snap := range changes {
					ld.nextSnapshot = snap
					ld.updateSnapshot()
				}
			}()
		}
	})
}

func NewMock(opts ...MockOption) *Loader {
	ld := new(Loader)
	for _, o := range opts {
		o.apply(ld)
	}
	var null loaderStats
	if ld.stats == null {
		ld.stats = newLoaderStats(stats.NewStore(mock.NewSink(), false))
	}
	ld.updateSnapshot()
	return ld
}

/*
type mockLoaderFunc struct {
	loaders map[string]map[string]*Loader
}

func (m *mockLoaderFunc) addLoader(path, subdir string, loader *Loader) {
	if m.loaders == nil {
		m.loaders = make(map[string]map[string]*Loader)
	}
	if m.loaders[path] == nil {
		m.loaders[path] = make(map[string]*Loader)
	}
	m.loaders[path][subdir] = loader
}

func (m *mockLoaderFunc) loadLoader(path, subdir string) *Loader {
	if sub := m.loaders[path]; sub != nil {
		return sub[subdir]
	}
	return nil
}

func MockNewFunc() func(string, string, stats.Scope, Refresher, ...Option) IFace {
	return New
}

func MockNew2Func() func(string, string, stats.Scope, Refresher, ...Option) (IFace, error) {
	return New2
}
*/

/*
func NewMock(snap *snapshot.Mock, changes <-chan *snapshot.Mock) (*Loader, *mock.Sink) {
	sink := mock.NewSink()
	ld := &Loader{
		nextSnapshot: snap,
		stats:        newLoaderStats(stats.NewStore(sink, false)),
	}
	ld.updateSnapshot()
	if changes != nil {
		go func() {
			for snap := range changes {
				ld.nextSnapshot = snap
				ld.updateSnapshot()
			}
		}()
	}
	return ld, nil
}
*/
