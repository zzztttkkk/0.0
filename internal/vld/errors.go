package vld

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/h2tp"
)

type ErrorReason int

const (
	ErrorReasonUndefined = ErrorReason(iota)
	ErrorReasonMissRequired
	ErrorReasonNumOutOfRange
	ErrorReasonLengthOutOfRange
	ErrorReasonNotMatchRegexp
	ErrorReasonBadTimeValue
	ErrorReasonCanNotCastToNum
	ErrorReasonCanNotCastToBool
)

var (
	errReasonStrings []string
)

func (er ErrorReason) String() string { return errReasonStrings[er] }

func (er ErrorReason) HttpStatus() int {
	switch er {
	case ErrorReasonUndefined:
		return h2tp.StatusInternalServerError
	default:
		return h2tp.StatusBadRequest
	}
}

func init() {
	errReasonStrings = []string{
		"Undefined",
		"MissRequired",
		"NumOutOfRange",
		"LengthOutOfRange",
		"NotMatchRegexp",
		"BadTimeValue",
		"CanNotCastToNum",
		"CanNotCastToBool",
	}
}

type Error struct {
	PkgPath  string
	TypeName string
	Reason   ErrorReason
	Rule     *Rule
	Input    any
}

func (err *Error) Detail() string {
	return fmt.Sprintf("%s %s.%s.%s %v", err.Reason, err.PkgPath, err.TypeName, err.Rule.Name, err.Input)
}

func (err *Error) Error() string { return fmt.Sprintf("%s %s", err.Reason, err.Rule.Name) }
