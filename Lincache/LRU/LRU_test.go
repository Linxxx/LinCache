package LRU

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestCache_Get(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || v.(String) != "1234" {
		t.Fatalf("cache hit key1=1234 fail")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 fail")
	}
}

func TestCache_Remove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Remove fail")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(16), callback)
	lru.Add("key1", String("1234"))
	lru.Add("key2", String("4321"))
	lru.Add("key3", String("5678"))
	lru.Add("key4", String("8765"))

	expect := []string{"key1", "key2"}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s, but %s",
			expect, keys)
	}
}
