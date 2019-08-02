package etterna

import "fmt"

func (e *Error) Error() string {
	if e.Msg == "" && e.Context == nil {
		panic("non-nil error with no details")
	}

	if e.Msg == "" {
		return e.Context.Error()
	} else if e.Context == nil {
		return e.Msg
	}

	return fmt.Sprintf("%s (%s)", e.Msg, e.Context.Error())
}

const (
	ErrUnexpected = iota
	ErrNotFound
)
