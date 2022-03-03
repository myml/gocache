package gocache

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	assert := require.New(t)
	cache, err := NewCache(10)
	assert.NoError(err)
	// 测试正常逻辑
	v, err := cache.Get("test", func() (interface{}, error) {
		return "hello", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")
	// 测试缓存逻辑
	v, err = cache.Get("test", func() (interface{}, error) {
		return "new", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")
	v, err = cache.Get("test", func() (interface{}, error) {
		return nil, errors.New("no_found")
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")

	// 测试错误不缓存
	v, err = cache.Get("no_found", func() (interface{}, error) {
		return nil, errors.New("no_found")
	}, time.Second, false)
	assert.Error(err)
	assert.Nil(v)
	v, err = cache.Get("no_found", func() (interface{}, error) {
		return "ok", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "ok")
	// 测试过期
	v, err = cache.Get("test_exp", func() (interface{}, error) {
		return "old", nil
	}, 0, false)
	assert.NoError(err)
	assert.Equal(v, "old")
	v, err = cache.Get("test_exp", func() (interface{}, error) {
		return "new", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "new")
}

func TestLru(t *testing.T) {
	assert := require.New(t)
	cache, err := NewCache(10)
	assert.NoError(err)
	// 测试键值淘汰
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		v, err := cache.Get(key, func() (interface{}, error) {
			return key, nil
		}, time.Hour, false)
		assert.NoError(err)
		assert.Equal(v, key)
	}
	assert.Equal(cache.store.Len(), 10)
}

func TestDelete(t *testing.T) {
	assert := require.New(t)
	cache, err := NewCache(10)
	assert.NoError(err)
	v, err := cache.Get("test", func() (interface{}, error) {
		return "hello", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")
	// 测试键值清理
	cache.Delete("test")
	v, err = cache.Get("test", func() (interface{}, error) {
		return nil, fmt.Errorf("not found")
	}, time.Second, false)
	assert.Error(err)
	assert.Nil(v)
}
func TestPurge(t *testing.T) {
	assert := require.New(t)
	cache, err := NewCache(10)
	assert.NoError(err)
	v, err := cache.Get("test1", func() (interface{}, error) {
		return "hello", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")
	v, err = cache.Get("test2", func() (interface{}, error) {
		return "hello", nil
	}, time.Second, false)
	assert.NoError(err)
	assert.Equal(v, "hello")
	// 测试清理
	cache.Clean()
	v, err = cache.Get("test1", func() (interface{}, error) {
		return nil, fmt.Errorf("not found")
	}, time.Second, false)
	assert.Error(err)
	assert.Nil(v)
	v, err = cache.Get("test2", func() (interface{}, error) {
		return nil, fmt.Errorf("not found")
	}, time.Second, false)
	assert.Error(err)
	assert.Nil(v)
}

func TestBackgroupUpdate(t *testing.T) {
	assert := require.New(t)
	cache, err := NewCache(10)
	assert.NoError(err)
	// 测试后台更新
	v, err := cache.Get("test_bg_update", func() (interface{}, error) {
		return "old", nil
	}, 0, true)
	assert.NoError(err)
	assert.Equal(v, "old")
	v, err = cache.Get("test_bg_update", func() (interface{}, error) {
		return "new", nil
	}, time.Second, true)
	assert.NoError(err)
	assert.Equal(v, "old")
	time.Sleep(time.Second / 100)
	v, err = cache.Get("test_bg_update", func() (interface{}, error) {
		return "new_new", nil
	}, time.Second, true)
	assert.NoError(err)
	assert.Equal(v, "new")
}
