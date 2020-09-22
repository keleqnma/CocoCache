package lru

import "container/list"

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	DefaultMaxBytes = 8192 * MEGABYTE
)

type Cache struct {
	// 允许使用的最大内存
	maxBytes int64
	// 当前已使用的内存
	usedBytes int64
	// 双向链表
	eleList *list.List
	//键是字符串，值是双向链表中对应节点的指针
	eleMap    map[string]*list.Element
	OnEvicted func(key string, value Value)
}

//双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

//用于返回值所占用的内存大小
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	cache := &Cache{
		maxBytes:  maxBytes,
		eleList:   list.New(),
		eleMap:    make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
	if maxBytes == 0 {
		cache.maxBytes = DefaultMaxBytes
	}
	return cache
}

func (c *Cache) Len() int {
	return c.eleList.Len()
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.eleMap[key]; ok {
		c.eleList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) Set(key string, value Value) {
	if ele, ok := c.eleMap[key]; ok {
		c.eleList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.usedBytes += int64(kv.value.Len()) - int64(value.Len())
		kv.value = value
	} else {
		ele := c.eleList.PushFront(&entry{key, value})
		c.eleMap[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.eleList.Back()
	if ele != nil {
		c.eleList.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.eleMap, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
	for c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}
