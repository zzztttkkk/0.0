package vld

type Vlder interface {
	FromString(string, *Error) bool
}
