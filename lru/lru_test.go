package lru

import (
	"errors"
	"fmt"
	"testing"
)

var (
	key1 = "key1"
	key2 = "key2"
	val1 = "julsj"
	val2 = "julusj"
)

type String string

func (d String) Len() int {
	return len(d)
}

func testSet(lru *Cache) {
	lru.Set(key1, String(val1))
	lru.Set(key2, String(val2))
}

func testGet(lru *Cache) error {
	if _, ok := lru.Get(key1); ok {
		return errors.New("无效的清除缓存")
	}
	value, ok := lru.Get(key2)
	if !ok {
		return errors.New("清除缓存错误")
	}
	if string(value.(String)) != val2 {
		return fmt.Errorf("设置缓存错误, expected:%v , got: %v", val2, string(value.(String)))
	}
	return nil
}

func TestLRU(t *testing.T) {
	lru := New(10, nil)
	testSet(lru)
	if err := testGet(lru); err != nil {
		t.Fatal(err)
	}
}
