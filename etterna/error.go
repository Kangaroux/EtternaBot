package etterna

func (e *Error) Error() string {
	if e.Msg == "" {
		return e.Context.Error()
	}

	return e.Msg
}

const (
	ErrUnexpected = iota
	ErrNotFound
)
