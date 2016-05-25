package keys

type contextKey int

const (
	// UserKey is the key to use when storing an entities.User into a context.
	UserKey contextKey = iota
)
