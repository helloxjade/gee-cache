package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)
//提供被其他节点访问的能力(基于http)
const defaultBasePath  = "/_geecache/"

type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self string     //用来记录自己的地址 包括主机名/ip
	basePath string   //作为节点间通讯地址的前缀  因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。比如，大部分网站的 API 接口，一般以 /api 作为前缀。
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

		w.Header().Set("Content-Type","application/octet-stream")
		w.Write(view.ByteSlice())
	}


