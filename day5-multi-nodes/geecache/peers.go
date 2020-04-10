package geecache
//抽象PeerPicker
// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string)(peer PeerGetter,ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string,key string)([]byte,error)  //从对应 group 查找缓存值。PeerGetter 就对应于上述流程中的 HTTP 客户端。
}



