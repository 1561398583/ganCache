package cache2

import (
	"ganCache/cache1"
	"strconv"
	"testing"
	"time"
)

//竞态检测
func TestCache2_ADD(t *testing.T) {
	c := &SafeCache{cacheBytes: 2<<10, cache: cache1.NewCache(2<<10, nil)}
	for i := 0; i < 10; i++{
		go writeCache2(c, strconv.FormatInt(int64(i), 10))
	}
	time.Sleep(time.Second * 20)
}


func writeCache2(c *SafeCache, name string)  {
	for i := 0; i < 10; i++ {
		k := name + strconv.FormatInt(int64(i), 10)
		v := "value" + strconv.FormatInt(int64(i), 10)
		c.ADD(k, []byte(v))
		time.Sleep(time.Second * 1)
	}
}

//数据安全检测
func TestSafeCache_ADD(t *testing.T) {
	c := &SafeCache{cacheBytes: 2<<10, cache: cache1.NewCache(2<<10, nil)}
	key := "test"
	value := []byte("abc")
	c.ADD(key, value)
	//修改value
	value[0] = 'b'
	//看cache中的"test"是否跟着变了
	value1, ok := c.Get(key)
	if !ok {
		t.Errorf("expect ok but not")
	}
	if string(value1) != "abc" {
		t.Errorf("expect \"abc\"  but %s", string(value1))
	}
}


