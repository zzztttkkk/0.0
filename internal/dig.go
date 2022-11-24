package internal

import (
	"go.uber.org/dig"
)

var c = dig.New()

func Provide(v any) {
	if err := c.Provide(v); err != nil {
		panic(err)
	}
}

func Invoke(fn any) {
	if err := c.Invoke(fn); err != nil {
		panic(err)
	}
}
