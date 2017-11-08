package snapshot

import (
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

// Mock provides a Snapshot implementation for testing
type Mock struct {
	*Snapshot
}

// NewMock initializes a new Mock
func NewMock() (s *Mock) {
	s = &Mock{
		Snapshot: New(),
	}

	return
}

// SetEnabled overrides the entry for `key` to be enabled
func (m *Mock) SetEnabled(key string) *Mock {
	m.Snapshot.entries[key] = &entry.Entry{
		StringValue: key,
		Uint64Value: 0,
		Uint64Valid: true,
		Modified:    time.Now(),
	}

	return m
}

// SetDisabled overrides the entry for `key` to be disabled
func (m *Mock) SetDisabled(key string) *Mock {
	m.Snapshot.entries[key] = &entry.Entry{
		StringValue: key,
		Uint64Value: 0,
		Uint64Valid: false,
		Modified:    time.Now(),
	}

	return m
}

// Set set the entry for `key` to `val`
func (m *Mock) Set(key string, val string) *Mock {
	m.Snapshot.entries[key] = &entry.Entry{
		StringValue: val,
		Uint64Value: 0,
		Uint64Valid: false,
		Modified:    time.Now(),
	}

	return m
}

// SetUInt64 set the entry for `key` to `val` as a uint64
func (m *Mock) SetUInt64(key string, val uint64) *Mock {
	m.Snapshot.entries[key] = &entry.Entry{
		StringValue: "",
		Uint64Value: val,
		Uint64Valid: true,
		Modified:    time.Now(),
	}

	return m
}

// FeatureEnabled overrides the internal `Snapshot`s `FeatureEnabled`
func (m *Mock) FeatureEnabled(key string, defaultValue uint64) bool {
	if e, ok := m.Snapshot.entries[key]; ok {
		return e.Uint64Valid
	}

	return false
}

var _ IFace = &Mock{}
