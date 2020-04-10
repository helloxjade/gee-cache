package geecache

import (
	"fmt"
	"gee-cache/day4-consistent-hash/geecache/consisitenthash"
	"github.com/golang/protobuf/proto"

	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	pb "gee-cache/day7-proto-buf/geecache/geecachepb"
)



//2.添加节点选择的功能
const(
	defaultBasePath="/_geecache/"
	defaultReplicas=50
)

type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self string     //用来记录自己的地址 包括主机名/ip
	basePath string   //作为节点间通讯地址的前缀  因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。比如，大部分网站的 API 接口，一般以 /api 作为前缀。
	mu sync.Mutex  //guards peers and httpGetters
	peers *consisitenthash.Map  //根据具体的key选择节点
	//映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关
	httpGetters map[string]*httpGetter   // keyed by e.g. "http://10.0.0.2:8008"

}

func NewHTTPPool(self string)*HTTPPool  {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}
// Log info with server name
func (p *HTTPPool) Log(format string,v ...interface{})  {
	log.Printf("[Server %s]%s",p.self,fmt.Sprintf(format,v...))
}

//ServeHTTP handle all http requests
func(p *HTTPPool) ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	if !strings.HasPrefix(r.URL.Path,p.basePath){
		panic("HTTPPool serving unexpected path:"+r.URL.Path)
	}
	p.Log("%s %s",r.Method,r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts:=strings.SplitN(r.URL.Path[len(p.basePath):],"/",2)
	if len(parts)!=2{
		http.Error(w,"bad request",http.StatusBadRequest)
		return
	}
	groupName:=parts[0]
	key:=parts[1]

	group:=GetGroup(groupName)
	if group==nil{
		http.Error(w,"no such group:"+groupName,http.StatusNotFound)
		return
	}
	//获得缓存实例
	view,err:=group.Get(key)
	if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	body,err:=proto.Marshal(&pb.Response{Value:view.ByteSlice()})
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(body)
}

//Set updates the pool's list of peers Set() 方法实例化了一致性哈希算法，并且添加了传入的节点
//并为每一个节点创建了一个 HTTP 客户端 httpGetter。
func(p *HTTPPool) Set(peers ...string)  {
     p.mu.Lock()
     defer p.mu.Unlock()
     p.peers=consisitenthash.New(defaultReplicas,nil)
     p.peers.Add(peers...)//peers 被打散传入
     p.httpGetters=make(map[string]*httpGetter,len(peers))
	for _,peer:=range peers{
		p.httpGetters[peer]=&httpGetter{baseURL:peer+p.basePath}
	}
}

// PickPeer picks a peer according to key
//ickerPeer() 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func(p *HTTPPool) PickPeer(key string) (PeerGetter,bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer:=p.peers.Get(key);peer!=""&&peer!=p.self{
		log.Printf("Pick peer %s",peer)
		return p.httpGetters[peer],true
	}
	return nil,false
}
var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	baseURL string  //http://example.com/_geecache/
}

func (h *httpGetter)Get(in *pb.Request,out *pb.Response)error{
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)

	//使用 http.Get() 方式获取返回值，并转换为 []bytes 类型
	res, err := http.Get(u)
	if err != nil {
		return  err
	}
	defer res.Body.Close()

	if res.StatusCode!=http.StatusOK{
		return fmt.Errorf("server returned :%v",res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err=proto.Unmarshal(bytes,out);err != nil {
		return fmt.Errorf("decoding response body :%v",err)
	}
	return nil
}
var _ PeerGetter=(*httpGetter)(nil)  //上面用来判断httpGetter是否实现了PeerGetter接口, 用作类型断言，如果httpGetter没有实现PeerGetter，则会报编译错误
