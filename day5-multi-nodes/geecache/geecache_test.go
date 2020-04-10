package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T)  {
	//我们借助 GetterFunc 的类型转换，将一个匿名回调函数转换成了接口 f Getter
	var f Getter=GetterFunc(func(key string)([]byte,error) {
	return []byte(key),nil
})
	expect:=[]byte("key")
	//调用该接口的方法 f.Get(key string)，实际上就是在调用匿名回调函数
	if v,_:=f.Get("key");!reflect.DeepEqual(v,expect){
	t.Errorf("callback failed")
	}
}



//1）在缓存为空的情况下，能够通过回调函数获取到源数据。
//2）在缓存已经存在的情况下，是否直接从缓存中获取，为了实现这一点，
// 使用 loadCounts 统计某个键调用回调函数的次数，如果次数大于1，则表示调用了多次回调函数，没有缓存。

var db=map[string]string{
	"Tom":"630",
	"Jack":"730",
	"Jerry":"830",
}
func TestGet(t *testing.T)  {
	loadCounts:=make(map[string]int,len(db))//调用回调函数的次数
	gee:=NewGroup("scores",2<<10,GetterFunc(func(key string)([]byte,error) {
		log.Println("[SlowDB] search key",key)
		if v,ok:=db[key];ok{
			if _,ok:=loadCounts[key];!ok{
				loadCounts[key]=0
			}
			loadCounts[key]+=1
			return []byte(v),nil
		}
		return nil,fmt.Errorf("%s not exist",key)
	}))
	for k,v:=range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {  //如果次数大于1，则表示调用了多次回调函数，没有缓存。
		t.Fatalf("cache %s miss", k)
		}

	}

	if view,err:=gee.Get("unknow");err==nil{
		t.Fatalf("the value of unknow should be empty,but %s got",view)
	}
}

