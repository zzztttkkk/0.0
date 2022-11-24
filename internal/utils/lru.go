package utils

import (
	"container/list"
	"time"
)

type LRUCache[K comparable, V any] struct {
	l       list.List
	m       map[K]*list.Element
	maxSize int
	maxAge  time.Duration
	zero    V
}

type _LRUCacheValue[K comparable, V any] struct {
	endAt int64
	key   K
	val   V
}

func NewLRUCache[K comparable, V any](maxSize int, maxAge time.Duration, zero V) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		m:       make(map[K]*list.Element),
		maxSize: maxSize,
		maxAge:  maxAge,
		zero:    zero,
	}
}

func (lruc *LRUCache[K, V]) Size() int { return len(lruc.m) }

func (lruc *LRUCache[K, V]) Store(key K, value V) {
	ele := lruc.m[key]
	if ele != nil {
		cv := ele.Value.(*_LRUCacheValue[K, V])
		cv.endAt = time.Now().Add(lruc.maxAge).UnixNano()
		cv.val = value
		lruc.l.MoveToFront(ele)
	} else {
		cv := &_LRUCacheValue[K, V]{
			endAt: time.Now().Add(lruc.maxAge).UnixNano(),
			key:   key,
			val:   value,
		}
		ele := lruc.l.PushFront(cv)
		lruc.m[key] = ele
	}

	for lruc.maxSize > 0 && len(lruc.m) > lruc.maxSize {
		back := lruc.l.Back()
		lruc.delEle(back)
	}
}

func (lruc *LRUCache[K, V]) delEle(ele *list.Element) {
	cv := ele.Value.(*_LRUCacheValue[K, V])
	delete(lruc.m, cv.key)
	lruc.l.Remove(ele)
}

func (lruc *LRUCache[K, V]) Del(key K) {
	ele := lruc.m[key]
	if ele == nil {
		return
	}
	lruc.delEle(ele)
}

func (lruc *LRUCache[K, V]) Load(key K) (V, bool) {
	ele := lruc.m[key]
	if ele == nil {
		return lruc.zero, false
	}

	cv := ele.Value.(*_LRUCacheValue[K, V])
	if lruc.maxAge > 0 && cv.endAt >= time.Now().UnixNano() {
		lruc.delEle(ele)
		return lruc.zero, false
	}
	lruc.l.MoveToFront(ele)
	return cv.val, true
}

func (lruc *LRUCache[K, V]) Reset() {
	lruc.l.Init()
	lruc.m = make(map[K]*list.Element)
}
