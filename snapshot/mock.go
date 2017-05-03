package snapshot

import "github.com/lyft/goruntime/snapshot/entry"

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
	m.Snapshot.entries[key] = entry.New(key, 0, true)

	return m
}

// SetDisabled overrides the entry for `key` to be disabled
func (m *Mock) SetDisabled(key string) *Mock {
	m.Snapshot.entries[key] = entry.New(key, 0, false)

	return m
}

// SetEntry set the entry for `key` to `val`
func (m *Mock) SetEntry(key string, val string) *Mock {
	m.Snapshot.entries[key] = entry.New(val, 0, false)

	return m
}

// FeatureEnabled overrides the internal `Snapshot`s `FeatureEnabled`
func (m *Mock) FeatureEnabled(key string, defaultValue uint64) bool {
	if e, ok := m.Snapshot.entries[key]; ok {
		return e.Uint64Valid
	}

	return false
}
