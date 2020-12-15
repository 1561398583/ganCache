package cache1

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)



func TestCache1_Get(t *testing.T) {
	cache := NewCache(100, nil)
	cache.Add("name", []byte("LiuBang"))
	v, ok := cache.Get("name")
	if !ok || string(v) != "LiuBang" {
		t.Errorf("Get => want : LiuBang, but not")
	}
	_, ok = cache.Get("name1")
	if ok  {
		t.Errorf("Get => want : nil, but not")
	}
}

//测试，当使用内存超过了设定值时，是否会触发“无用”节点的移除
func TestCache1_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + v1 + k2 + v2)
	cache := NewCache(int64(cap), nil)
	cache.Add(k1, []byte(v1))
	cache.Add(k2, []byte(v2))
	cache.Add(k3, []byte(v3))

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("RemoveOldest  key1 fail")
	}
}

//测试回调函数能否被调用
func TestCache1_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value []byte) {
		keys = append(keys, key)
	}
	k1, k2, k3, k4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cap1 := len(k1 + v1 + k2 + v2)
	cache := NewCache(int64(cap1), callback)
	cache.Add(k1, []byte(v1))
	cache.Add(k2, []byte(v2))
	cache.Add(k3, []byte(v3))
	cache.Add(k4, []byte(v4))

	expct := []string{"key1", "key2"}
	if !reflect.DeepEqual(keys, expct) {
		t.Errorf("call OnEvicted failed, expct keys equels to %s", expct)
	}
}


//竞态检测
func TestCache1_Get2(t *testing.T) {
	cache := NewCache(2<<10, nil)
	for i := 0; i < 10; i++{
		go writeCache1(cache, strconv.FormatInt(int64(i), 10))
	}
	time.Sleep(time.Second * 20)
}


func writeCache1(c *Cache, name string)  {
	for i := 0; i < 10; i++ {
		k := name + strconv.FormatInt(int64(i), 10)
		v := "value" + strconv.FormatInt(int64(i), 10)
		c.Add(k, []byte(v))
		time.Sleep(time.Second * 1)
	}
}


