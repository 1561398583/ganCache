package cache3

import (
	"ganCache/cache2"
	"github.com/pkg/errors"
)

//在CacheWithMutex的基础上再封装一层，
//支持用户自定义Find，Find的功能为当cache中没找到数据时，去哪里获取数据（文件、数据库、网络...）
type CacheWithGetter struct {
	//并发安全的数据容器
	c *cache2.SafeCache
	//用户自定义的getter，当缓存不存在时的回调函数
	finder	Finder
}

// A Getter loads data for a key.
type Finder interface {
	Find(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type FindFunc func(key string) ([]byte, error)

// GetterFunc实现Getter interface
func (f FindFunc) Find(key string) ([]byte, error) {
	return f(key)
}

func (g *CacheWithGetter) Add(key string, value []byte){
	g.c.ADD(key, value)
}

func (g *CacheWithGetter) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("key can not be empty")
	}
	data, ok := g.c.Get(key)
	if ok {
		return data, nil
	}
	//缓存中未找到，于是调用用户自定义的finder从源获取数据
	if g.finder == nil {
		return nil, nil
	}
	bs, err := g.finder.Find(key)
	if err != nil {
		return nil, err
	}
	//加入cache
	g.c.ADD(key, bs)
	//为了避免外部程序修改cache中的数据，这里返回一个copy
	rb := make([]byte, len(bs))
	copy(rb, bs)
	return rb, nil
}


func NewCacheWithGetter(cacheSize int64, finder Finder) *CacheWithGetter {
	cg := &CacheWithGetter{
		c:      cache2.NewSafeCache(cacheSize),
		finder: finder,
	}

	return cg
}

