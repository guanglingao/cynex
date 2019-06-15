package cache

import (
	"fmt"
	"strconv"
	"testing"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache(1)
	cache.Set("key", "this is zhe vale")
	if c, _ := cache.Get("key"); c != "this is zhe vale" {
		t.Fail()
	}
}

func TestCache_Set(t *testing.T) {
	cache := NewCache(10)
	for i := 0; i < 100; i++ {
		cache.Set("key"+strconv.Itoa(i), "val"+strconv.Itoa(i))
	}
	fmt.Println(cache.hashData)
	cache = NewCache()
	for i := 0; i < 343; i++ {
		cache.Set("key"+strconv.Itoa(i), "val"+strconv.Itoa(i))
	}
	fmt.Println(cache.hashData)
}
