package cache1

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	//cache最大使用量
	maxBytes int64
	//cache当前使用量
	nBytes int64
	//存放数据的容器
	cache map[string]*list.Element
	//数据的排序
	ll *list.List
	//当key被删除后的回调函数，可以为nil
	onEvicted func(key string, value []byte)
}

type entry struct {
	key   string
	//[]byte类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
	value []byte
}


func NewCache(maxBytes int64, onEvicted func(key string, value []byte)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		onEvicted: onEvicted,
		nBytes: 0,
		cache: make(map[string]*list.Element),
		ll: list.New(),
	}
}

//1.加入map；2.把新加入的移到ll的头部
func (c *Cache) Add(key string, value []byte)  {
	if ele, ok := c.cache[key]; ok{	//如果已经存在
		//*Element移到头部
		c.ll.MoveToFront(ele)
		//获取Element中的entry
		ent := ele.Value.(*entry)
		//计算新的容量
		c.nBytes = c.nBytes + int64(len(value)) - int64(len(ent.value))
		//entry的value赋值为新value
		ent.value = value
	}else {
		//先把key-value封装进entry，然后把entry封装进Element并加到ll的头部，返回*Element
		ele := c.ll.PushFront(&entry{key: key, value: value})
		//*Element放入cache
		c.cache[key] = ele
		//计算新的容量
		c.nBytes = c.nBytes + int64(len(key)) + int64(len(value))
	}

	//已使用的内存超过最大值，删除一些，直到已使用的值小于最大值
	for c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

//查找主要有 2 个步骤，第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾。
func (c *Cache) Get(key string) (value []byte, ok bool) {
	if ele, ok := c.cache[key]; ok {
		//移到ll头部
		c.ll.MoveToFront(ele)
		ent := ele.Value.(*entry)
		return ent.value, true
	}else {
		return nil, false
	}
}

func (c *Cache) RemoveOldest()  {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	//从ll中删除
	c.ll.Remove(ele)
	ent := ele.Value.(*entry)
	//从cache中删除
	delete(c.cache, ent.key)
	//计算新的容量
	c.nBytes = c.nBytes - int64(len(ent.key)) - int64(len(ent.value))
	//回调函数
	if c.onEvicted != nil {
		c.onEvicted(ent.key, ent.value)
	}
}

// 便于测试
func (c *Cache) Len() int {
	return c.ll.Len()
}


