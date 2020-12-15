package cache2

import (
	"ganCache/cache1"
	"sync"
)

//给Cache加一层,增加并发安全和数据安全
//加了锁，可以并发安全
//ADD和Get的数据都进行copy，这样就能防止外部程序修改cache中的数据
type SafeCache struct {
	//同步锁，保证并发安全
	mu sync.Mutex
	//存放数据的容器
	cache *cache1.Cache
	cacheBytes int64
}

func NewSafeCache(cacheSize int64)  *SafeCache {
	return &SafeCache{cache: cache1.NewCache(cacheSize, nil), cacheBytes: cacheSize}
}

//默认数据缓存容器
//参数不一定要传过来，也可以从固定的地方去取。
//这里就是从静态变量defaultCache（编译后的固定地址）中取得一个*cache。
var defaultCache *SafeCache

func DefaultCache()  *SafeCache {
	if defaultCache == nil {
		defaultCache = &SafeCache{cache: cache1.NewCache(2<<10, nil), cacheBytes: 2<<10}
	}
	return defaultCache
}

func (c *SafeCache) ADD(key string, value []byte)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	//延迟初始化，等于 nil 再创建实例。主要用于提高性能，并减少程序内存要求。
	if c.cache == nil {
		c.cache = cache1.NewCache(c.cacheBytes, nil)
	}
	//因为value是从外部程序传过来的，虽然value是外部程序的slice的一个副本，但是都指向同一块内存空间。
	//如果直接把value存入cache中，那么外部程序可以通过value原本slice修改cache中的数据。
	//所以这里value需要copy再存入cache。
	cv := make([]byte, len(value))	//在内存中新开辟一块内存空间
	copy(cv, value)
	c.cache.Add(key, cv)
}

func (c *SafeCache) Get(key string)  (value []byte, ok bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	v, ok := c.cache.Get(key)	//返回lru.Value接口
	if !ok {
		return
	}

	//同样是为了避免外部程序修改cache中的数据，这里返回v的copy
	r := make([]byte, len(v))
	copy(r, v)
	return r, true
}