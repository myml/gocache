package gocache

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/sync/singleflight"
)

type cacheEntrie struct {
	ExpireTime time.Time
	Value      interface{}
}

type expireKeyType string

// Cache 一个内存缓存库，支持过期和后台更新
type Cache struct {
	single singleflight.Group
	store  *lru.ARCCache
}

// NewCache 初始化缓存，缓存数量超过size将使用lru算法淘汰
func NewCache(size int) (*Cache, error) {
	arc, err := lru.NewARC(size)
	if err != nil {
		return nil, err
	}
	return &Cache{store: arc}, nil
}

// Get 获取key缓存的值，如不存在则使用loader函数加载缓存
//
// 如果使用后台更新，函数直接返回过期的值，并使用协程调用loader更新缓存
func (c *Cache) Get(key string, loader func() (interface{}, error), expire time.Duration, backgroupUpdate bool) (interface{}, error) {
	v, err, _ := c.single.Do(key, func() (interface{}, error) {
		load := func() (interface{}, error) {
			v, err := loader()
			if err != nil {
				return nil, err
			}
			entrie := &cacheEntrie{ExpireTime: time.Now().Add(expire), Value: v}
			c.store.Add(key, entrie)
			return entrie.Value, nil
		}
		v, ok := c.store.Get(key)
		// 初始化
		if !ok {
			return load()
		}
		entrie := v.(*cacheEntrie)
		// 未过期
		if time.Since(entrie.ExpireTime) < 0 {
			return entrie.Value, nil
		}
		// 过期更新
		if !backgroupUpdate {
			return load()
		}
		// 过期后台更新
		go func() {
			c.single.Do(key+"_backgroup_update", func() (interface{}, error) {
				return load()
			})
		}()
		return entrie.Value, nil
	})
	return v, err
}

// Delete 删除指定键缓存
func (c *Cache) Delete(key string) {
	c.store.Remove(key)
	c.store.Remove(expireKeyType(key))
}

// Clean 清理所有缓存内容
func (c *Cache) Clean() {
	c.store.Purge()
}
