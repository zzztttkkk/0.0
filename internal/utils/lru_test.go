package utils

import (
	"fmt"
	"testing"
)

func TestLRUCacheSL(t *testing.T) {
	c := NewLRUCache[string, int](10, 0, 0)
	c.Store("X", 45)
	c.Store("X", 46)
	v, ok := c.Load("X")
	if !ok || v != 46 {
		t.Fail()
	}
}

func TestLRUCacheMaxSize(t *testing.T) {
	c := NewLRUCache[string, int](10, 0, 0)
	for i := 0; i < 100; i++ {
		c.Store(fmt.Sprintf("%d", i), i)
	}
	if c.Size() != 10 {
		t.Fail()
	}
}
