package goocache

import (
	"fmt"
	"log"
	"sync"
)

// Group是一个缓存命名空间和相关数据加载功能的集合
// 每个Group都有一个唯一的名字name和一个缓存getter，mainCache是一个线程安全的lru.Cache
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

// Getter是一个接口，定义了一个Get方法，用于根据key获取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc是一个函数类型，实现了Getter接口的Get方法，使得普通函数也能作为数据获取的方式
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup创建一个新的Group实例，并将其添加到全局的groups映射中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup返回之前通过NewGroup创建的指定名称的Group实例，如果没有找到则返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get从缓存中获取key对应的值，如果缓存中没有，则调用load方法从数据源加载数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// load方法从数据源加载数据，并将加载的数据添加到缓存中
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// getLocally方法使用getter从数据源获取数据，并将获取的数据添加到缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache是一个辅助方法，用于将从数据源获取的数据添加到缓存中，以便下次可以直接从缓存中获取
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
