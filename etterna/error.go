package etterna

import "fmt"

func (e *Error) Error() string {
	if e.Msg == "" {
		return e.Context.Error()
	}

	return fmt.Sprintf("%s (%s)", e.Msg, e.Context)
}

const (
	ErrUnexpected = iota
	ErrNotFound
)
