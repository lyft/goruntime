package snapshot

import (
	"time"

	"github.com/lyft/goruntime/snapshot/entry"
)

// Snapshot provides the currently loaded set of runtime values.
type IFace interface {
	// Test if a feature is enabled using the built in random generator. This is done by generating
	// a random number in the range 0-99 and seeing if this number is < the value stored in the
	// runtime key, or the default_value if the runtime key is invalid.
	//
	// NOTE: Although a snapshot represents a stable snapshot *of the contents in runtime*, the behavior of
	//       this function, by design, includes performing a pseudo-random dice roll upon every call to it.
	//       As a result, despite the fact that the underlying snapshot is not changing, the result of calling
	//       this function repeatedly with the same parameter will *not* necessarily yield the same result.
	//       Callers must be careful not to assume that multiple calls with result in a consistent return value.
	//
	//       In other words, the snapshot provides a fixed *probability* of a particular result, but the result
	//       will still vary across calls based on that probability.
	//
	// @param key supplies the feature key to lookup.
	// @param defaultValue supplies the default value that will be used if either the feature key
	//        does not exist or it is not an integer.
	// @return true if the feature is enabled.
	FeatureEnabled(key string, defaultValue uint64) bool

	// FeatureEnabledForID checks that the crc32 of the id and key's byte value falls within the mod of
	// the 0-100 value for the given feature. Use this method for "sticky" features
	// @param key supplies the feature key to lookup.
	// @param id supplies the ID to use in the CRC check.
	// @param defaultValue supplies the default value that will be used if either the feature key
	//        does not exist or it is not a valid percentage.
	FeatureEnabledForID(key string, id uint64, defaultPercentage uint32) bool

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

	// GetModified returns the last modified timestamp for key. If key does not
	// exist, the zero value for time.Time is returned.
	GetModified(key string) time.Time

	// Fetch all keys inside the snapshot.
	// @return []string all of the keys.
	Keys() []string

	Entries() map[string]*entry.Entry

	SetEntry(string, *entry.Entry)
}
