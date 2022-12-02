package errors

type Error struct {
	Code   ErrorCode
	Msg    string
	Detail any
}

func (e *Error) Error() string {
	return codeToString(e.Code)
}
