package streamlive

type Resolver interface {
	Next() error

	nextFunc() func() error
	setNextFunc(f func() error)
}

type Event struct {
	next func() error
}

// Next calls the next hook handler.
func (e *Event) Next() error {
	if e.next != nil {
		return e.next()
	}
	return nil
}

// nextFunc returns the function that Next calls.
func (e *Event) nextFunc() func() error {
	return e.next
}

// setNextFunc sets the function that Next calls.
func (e *Event) setNextFunc(fn func() error) {
	e.next = fn
}
