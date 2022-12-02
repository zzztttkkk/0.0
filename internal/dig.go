package internal

import (
	"context"
	"fmt"
	"go.uber.org/dig"
	"reflect"
	"sync"
	"time"
)

var c = dig.New()

var (
	lock     sync.Mutex
	invokes  = make(map[uintptr]any)
	provides = make(map[uintptr]any)
)

func consume(k uintptr, v any, fn func(any) error, consumes *[]uintptr) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err := fn(v); err == nil {
		*consumes = append(*consumes, k)
	}
}

func InvokeAll(timeout int) <-chan struct{} {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	var ch = make(chan struct{}, 1)

	go func() {
		var consumedInvokes []uintptr
		var consumedProvides []uintptr
		var stop bool

		for !stop {
			select {
			case <-ctx.Done():
				stop = true
				break
			default:
				{
					consumedInvokes = consumedInvokes[:0]
					consumedProvides = consumedProvides[:0]

					lock.Lock()

					for ptr, fn := range provides {
						consume(ptr, fn, func(a any) error { return c.Provide(a) }, &consumedProvides)
					}

					for ptr, fn := range invokes {
						consume(ptr, fn, func(a any) error { return c.Invoke(a) }, &consumedInvokes)
					}

					for _, v := range consumedProvides {
						delete(provides, v)
					}
					for _, v := range consumedInvokes {
						delete(invokes, v)
					}

					lock.Unlock()

					time.Sleep(time.Millisecond * 30)
				}
			}
		}

		ch <- struct{}{}

		for _, fn := range invokes {
			if err := c.Invoke(fn); err != nil {
				panic(err)
			}
		}
	}()

	return ch
}

func LazyInvoke(fn any) {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic(fmt.Errorf(`%s is not function`, fn))
	}

	lock.Lock()
	defer lock.Unlock()

	invokes[reflect.ValueOf(fn).Pointer()] = fn
}

func LazyProvide(fn any) {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic(fmt.Errorf(`%s is not function`, fn))
	}

	lock.Lock()
	defer lock.Unlock()

	provides[reflect.ValueOf(fn).Pointer()] = fn
}

func Invoke(fn any) {
	if err := c.Invoke(fn); err != nil {
		fmt.Println(err)
	}
}

func Provide(v any) {
	if err := c.Provide(v); err != nil {
		panic(err)
	}
}
