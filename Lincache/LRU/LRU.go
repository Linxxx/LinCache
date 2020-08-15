/*
	实现LRU缓存
*/
package LRU

import (
	"container/list"
)

type Cache struct {
	maxMemBytes int64 // 最大可使用内存大小
	tmpMemBytes int64 // 当前已使用内存大小

	deLinkedList *list.List                    // 双向链表
	cache        map[string]*list.Element      // 哈希表索引指向链表节点的指针
	OnEvicted    func(key string, value Value) // 回调函数，可为nil
}

type node struct {
	// 表示key-value键值对
	key   string
	value Value
}

type Value interface {
	// 这个接口用于计算当前缓存占用了多少字节
	Len() int
}

func (c *Cache) Len() int {
	// 便于获取cache大小
	return c.deLinkedList.Len()
}

func New(maxMemBytes int64, OnEvicted func(string, Value)) *Cache {
	// 构造函数
	return &Cache{
		maxMemBytes:  maxMemBytes,
		deLinkedList: list.New(),
		cache:        make(map[string]*list.Element),
		OnEvicted:    OnEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	// 访问某个节点
	if element, ok := c.cache[key]; ok {
		c.deLinkedList.MoveToFront(element)
		kv := element.Value.(*node)
		return kv.value, true
	}
	return
}

func (c *Cache) Remove() {
	// 删除最长最久未使用的节点
	element := c.deLinkedList.Back()
	if element != nil {
		c.deLinkedList.Remove(element)
		kv := element.Value.(*node)
		delete(c.cache, kv.key)
		c.tmpMemBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// 新增/更新某个节点
	if element, ok := c.cache[key]; ok {
		c.deLinkedList.MoveToFront(element)
		kv := element.Value.(*node)
		c.tmpMemBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		element := c.deLinkedList.PushFront(&node{key, value})
		c.cache[key] = element
		c.tmpMemBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxMemBytes != 0 && c.maxMemBytes < c.tmpMemBytes {
		c.Remove()
	}
}
