package entry

import "time"

// An individual snapshot entry. Optimized for integers by pre-converting them if possible.
type Entry struct {
	StringValue string
	Uint64Value uint64
	Uint64Valid bool
	Modified    time.Time
}
