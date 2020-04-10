package geecache

import (
	"fmt"
	"log"
	"sync"
)
//负责与外部交互，控制缓存存储和获取的主流程
//A Getter loads data for a key 加载键的数值
type Getter interface {
	Get(key string)([]byte,error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string)([]byte,error)   //回调函数


// Get implements Getter interface function  //实现Getter的方法其实是在调用自身
func(f GetterFunc) Get(key string)([]byte,error)  {
	return f(key)
}




// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name string
	getter Getter  //缓存未命中时获取源数据的回调(callback)
	mainCache cache   //一开始实现的并发缓存
}

var (
	mu sync.RWMutex
	groups=make(map[string]*Group)  //用来存储Group的全局变量
)

//NewGroup create a new instance of Group
func NewGroup(name string,cacheBytes int64,getter Getter)*Group  {
	if getter==nil{
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g:=&Group{
		name:name,
		getter:getter,
		mainCache:cache{cacheBytes:cacheBytes},
	}
	groups[name]=g
	return g

}

//GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GeyGroup(name string)*Group  {//获取特定名称的group
	mu.RLock()  //不涉及任何冲突变量的写操作
	g:=groups[name]
	mu.RUnlock()
	return g
}

//Get value for a key from cache
func (g *Group)Get(key string)(ByteView,error)  {
	if key==""{
		return ByteView{},fmt.Errorf("key is required")
	}
	if v,ok:=g.mainCache.get(key);ok{
		log.Println("[GeeCache] hit")
		return v,nil
	}
	return g.load(key)
}
func(g *Group) load(key string)(value ByteView,err error) {
     return g.getLocally(key)	//分布式场景下会调用 getFromPeer 从其他节点获取
}

func(g *Group) getLocally(key string)(ByteView, error)  {
	bytes,err:=g.getter.Get(key)//调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
	if err!=nil{
		return ByteView{},err
	}
	value :=ByteView{b:cloneBytes(bytes)}
	g.populateCache(key,value)
	return value,nil
}

func(g *Group) populateCache(key string,value ByteView)  {
	g.mainCache.add(key,value)
}