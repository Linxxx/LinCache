package Lincache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":   "444",
	"Kate":  "589",
	"Linda": "312",
	"Sam":   "325",
}

func TestGetterFunc_Get(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

func TestGetGroup(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	linx := NewGroup("No", 1<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		if view, err := linx.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		}
		if _, err := linx.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}
}
