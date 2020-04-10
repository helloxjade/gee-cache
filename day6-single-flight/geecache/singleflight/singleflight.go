package singleflight

import "sync"

type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

//管理不同的key的请求
type Group struct {
	mu sync.Mutex
	m map[string]*call
}


//针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误
func(g *Group) Do(key string,fn func()(interface{},error)) (interface{},error) {
	g.mu.Lock()
	if g.m==nil{
		g.m=make(map[string]*call)  //延迟初始化
	}
	if c,ok:=g.m[key];ok{
		g.mu.Unlock()
		c.wg.Wait()  //如果请求在进行中，则等待
		return c.val,c.err
	}
	c:=new(call)
	c.wg.Add(1)     // 发起请求前加锁
	g.m[key]=c   //表明key已经有对应的请求处理
	g.mu.Unlock()

	c.val,c.err=fn()    //调用fn 发起请求
	c.wg.Done()  //请求结束

	g.mu.Lock()
	delete(g.m,key)    //更新g.m
	g.mu.Unlock()  //锁减一
	//并发协程之间不需要消息传递，非常适合 sync.WaitGroup
	return c.val,c.err
}
