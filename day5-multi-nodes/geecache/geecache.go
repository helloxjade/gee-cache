package geecache

import (
	"fmt"
	"log"
	"sync"
)

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name string
	getter Getter
	mainCache cache
	peers PeerPicker

}
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
func GetGroup(name string)*Group  {//获取特定名称的group
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

//// RegisterPeers registers a PeerPicker for choosing remote peer
func(g *Group) RegisterPeers(peers PeerPicker)  {
	if g.peers!=nil{
		panic("RegisterPeerPicker called more than once")
	}
	g.peers=peers  // PeerPicker 接口的 HTTPPool 注入到 Group 中
}


func(g *Group) load(key string)(value ByteView,err error)  {
	if g.peers!=nil{
		if peer,ok:=g.peers.PickPeer(key);ok{
			if value,err=g.getFromPeer(peer,key);err==nil{
				return value,nil
			}
			log.Println("[GeeCache] failed to get from peer",err)
		}
	}
	return g.getLocally(key)
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

func (g *Group)getFromPeer(peer PeerGetter,key string)(ByteView,error) {
     bytes,err:=peer.Get(g.name,key)
	if err != nil {
		return ByteView{},err
	}
	return ByteView{b:bytes},nil
}
