package gocache

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type expireKeyType string

// Cache 一个内存缓存库，支持过期和后台更新
type Cache struct {
	single singleflight.Group
	store  sync.Map
}

// Get 获取key缓存的值，如不存在则使用loader函数加载缓存
//
// 如果使用后台更新，函数直接返回过期的值，并使用协程调用loader更新缓存
func (c *Cache) Get(key string, loader func() (interface{}, error), expire time.Duration, backgroupUpdate bool) (interface{}, error) {
	v, err, _ := c.single.Do(key, func() (interface{}, error) {
		expKye := expireKeyType(key)
		load := func() (interface{}, error) {
			v, err := loader()
			if err != nil {
				return nil, err
			}
			c.store.Store(key, v)
			c.store.Store(expKye, time.Now().Add(expire))
			return v, nil
		}
		exp, ok := c.store.Load(expKye)
		// 初始化
		if !ok {
			return load()
		}
		// 未过期
		t := exp.(time.Time)
		if time.Since(t) < 0 {
			v, _ := c.store.Load(key)
			return v, nil
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
		v, _ := c.store.Load(key)
		return v, nil
	})
	return v, err
}

// Delete 清理缓存
func (c *Cache) Delete(key string) {
	c.store.Delete(key)
	c.store.Delete(expireKeyType(key))
}
