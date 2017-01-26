package entry

// An individual snapshot entry. Optimized for integers by pre-converting them if possible.
type Entry struct {
	StringValue string
	Uint64Value uint64
	Uint64Valid bool
}

func New(s string, ui uint64, v bool) (e *Entry) {
	e = &Entry{s, ui, v}

	return
}
