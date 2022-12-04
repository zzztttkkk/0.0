package vld

import (
	"reflect"
	"testing"
)

type Address struct {
	City string `vld:"city"`
}

type User struct {
	Name    string  `vld:"name"`
	Email   string  `vld:"email"`
	Age     int     `vld:"age"`
	Address Address `vld:"address"`
	Friends []User  `vld:"friends"`
}

func TestGetRules(t *testing.T) {
	GetRules(reflect.TypeOf(User{}))
}
