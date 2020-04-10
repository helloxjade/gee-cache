package geecache

import (
	"gee-cache/day2-single-node/geecache/lru"
	"sync"
)
//为 lru.cache添加并发特性

type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}
//实例化lru  封装add和get方法
func(c *cache) add(key string,value ByteView)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		//如果不等于 nil 再创建实例。这种方法称之为延迟初始化(Lazy Initialization)，
		// 一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时  可以提高性能，减少程序内存要求
		c.lru=lru.New(c.cacheBytes,nil)
	}
	c.lru.Add(key,value)
}

func (c *cache)get(key string)(value ByteView,ok bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		return
	}
	if v,ok:=c.lru.Get(key);ok{
		return v.(ByteView),ok
	}
	return
}
