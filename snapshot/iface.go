package snapshot

import "github.com/lyft/goruntime/snapshot/entry"

// Snapshot provides the currently loaded set of runtime values.
type IFace interface {
	// Test if a feature is enabled using the built in random generator. This is done by generating
	// a random number in the range 0-99 and seeing if this number is < the value stored in the
	// runtime key, or the default_value if the runtime key is invalid.
	// NOTE: In the current implementation, this routine may return different results each time it
	//       is called because a new random number is used each time. Callers should understand this
	//       behavior and not assume that subsequent calls using the same snapshot will be
	//       consistent.
	// @param key supplies the feature key to lookup.
	// @param defaultValue supplies the default value that will be used if either the feature key
	//        does not exist or it is not an integer.
	// @return true if the feature is enabled.
	FeatureEnabled(key string, defaultValue uint64) bool

	// Fetch raw runtime data based on key.
	// @param key supplies the key to fetch.
	// @return const std::string& the value or empty string if the key does not exist.
	Get(key string) string

	// Fetch an integer runtime key.
	// @param key supplies the key to fetch.
	// @param defaultValue supplies the value to return if the key does not exist or it does not
	//        contain an integer.
	// @return uint64 the runtime value or the default value.
	GetInteger(key string, defaultValue uint64) uint64

	// Fetch all keys inside the snapshot.
	// @return []string all of the keys.
	Keys() []string

	Entries() map[string]*entry.Entry

	SetEntry(string, *entry.Entry)
}
