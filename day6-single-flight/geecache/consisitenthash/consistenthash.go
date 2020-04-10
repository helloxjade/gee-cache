package consisitenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash Hash     //Hash 函数
	replicas int    //虚拟节点的倍数
	keys []int  //sorted  哈希环
	hashMap map[int]string  //虚拟节点与真实节点的映射表//  键是虚拟节点的哈希值，值是真实节点的名称。
}

//允许自定义虚拟节点倍数和 Hash 函数
func New(replicas int, fn Hash)*Map  {
	m:=&Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash==nil{
		m.hash=crc32.ChecksumIEEE
	}
	return m
}


//允许传入0或多个真实节点的名称
func(m *Map) Add(keys ...string)  {
	for _,key :=range keys{
		//每个真实节点key 对应创建m.repliacas个虚拟节点
		for i:=0;i<m.replicas;i++{
			//虚拟节点的名称为strconv.Itoa(i)+key  即通过添加不同的编号区分不同的虚拟节点
			hash:=int(m.hash([]byte(strconv.Itoa(i)+key)))  //计算虚拟节点的哈希值
			m.keys=append(m.keys,hash) //添加到环上
			m.hashMap[hash]=key //增加虚拟节点和真实节点的映射关系
		}
	}
	sort.Ints(m.keys) //环上的哈希值排序
}


//选择节点 的get方法
func(m *Map) Get(key string) string {
	if len(m.keys)==0{
		return ""
	}
	hash:=int(m.hash([]byte(key)))
	idx:=sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i]>=hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}