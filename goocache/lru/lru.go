package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 最大内存限制
	nowBytes  int64                         // 当前内存使用
	ll        *list.List                    // 双向链表，保存key和value
	cache     map[string]*list.Element      // key到双向链表中节点的映射
	onEvicted func(key string, value Value) // 可在被移除时执行，作为回调函数（比如记录日志）
}

// entry是双向链表中节点保存的值
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nowBytes += int64(value.Len()) - int64(kv.value.Len()) //新增的value占用内存 - 原来value占用内存
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nowBytes += int64(len(key)) + int64(value.Len()) //新增的key占用内存 + 新增的value占用内存
	}
	for c.maxBytes != 0 && c.maxBytes < c.nowBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nowBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)

		}
	}
}
func (c *Cache) Len() int {
	return c.ll.Len()
}
