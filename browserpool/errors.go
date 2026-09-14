package browserpool

import pool "github.com/ToufiqQureshi/Scraper/internal/pool"

// ErrClosed is returned when Get is called on a pool that has already
// been closed.
var ErrClosed = pool.ErrClosed

// ErrInvalidSize is returned by New when size is not positive.
var ErrInvalidSize = pool.ErrInvalidSize

// Error reports what operation failed and why, so callers can tell a
// closed pool apart from a Chrome launch failure without parsing text.
type Error struct {
	Op  string // "new" or "get"
	Err error
}

func (e *Error) Error() string {
	return "browserpool: " + e.Op + ": " + e.Err.Error()
}

// Unwrap lets errors.Is(err, browserpool.ErrClosed) work through this type.
func (e *Error) Unwrap() error {
	return e.Err
}
