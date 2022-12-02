package errors

type ErrorCode int

const (
	// BadParamsError http: 400
	BadParamsError = ErrorCode(iota)
)

var (
	codeToString func(ErrorCode) string
)
