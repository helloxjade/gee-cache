package lru

import "container/list"

type Cache struct {
	maxBytes int64  //允许使用的最大内存
	nbytes int64  //当前已使用的内存
	ll   *list.List //Go语言标准库实现的双向链表
	cache map[string]*list.Element //值是双向链表中对应节点的指针
	// optional and executed when an entry is purged.
	OnEvicted func(key string ,value Value)//某条记录被移除时的回调函数，可以为nil
}
//双向链表节点的数据类型 ,保存key的原因是淘汰方便
type entry struct {
	key string
	value Value
}


//Value use Len to count how many bytes it takes
type Value interface {
	Len() int                 //返回值所占用的内存大小
}

//实例化Cache
func New(maxBytes int64,onEvicted func(string,Value)) *Cache  {
	return &Cache{
		maxBytes:maxBytes,
		ll:list.New(),
		cache:make(map[string]*list.Element),
		OnEvicted:onEvicted,
		}
}

//查找功能
func (c *Cache)Get(key string) (value Value,ok bool) {
	if ele ,ok :=c.cache[key];ok{   //如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
		c.ll.MoveToFront(ele)       //双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾
		kv:=ele.Value.(*entry)
		return kv.value,true
	}
	return
}
//// RemoveOldest removes the oldest item
func (c *Cache)RemoveOldest()  {
	ele:=c.ll.Back()   //回到队首节点，从链表中删除
	if ele!=nil{
		c.ll.Remove(ele)
		kv:=ele.Value.(*entry)
		delete(c.cache,kv.key)   //从字典中删除该节点的映射关系
		c.nbytes-=int64(len(kv.key))+int64(kv.value.Len())  //更新当前所用的内存
         if c.OnEvicted!=nil{     //如果回调函数不为空，则调用回调函数
         	c.OnEvicted(kv.key,kv.value)
		 }
	}
}


// Add adds a value to the cache.
func(c *Cache) Add(key string,value Value)  {
	if ele,ok :=c.cache[key];ok{
		c.ll.MoveToFront(ele)   //如果键存在，则更新对应节点的值，并将该节点移到队尾
		kv:=ele.Value.(*entry)
		c.nbytes+=int64(value.Len())-int64(kv.value.Len())
		kv.value=value
		return
	}
	ele:=c.ll.PushFront(&entry{key,value})  //不存在在队尾添加新节点
	c.cache[key]=ele  //在字典中添加key和新节点的映射关系
	c.nbytes+=int64(len(key))+int64(value.Len())
	for c.maxBytes !=0&&c.maxBytes<c.nbytes{  //如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点
		c.RemoveOldest()
	}
}

//为了方便测试，我们实现 Len() 用来获取添加了多少条数据
func(c *Cache)Len()int{
	return c.ll.Len()
}

