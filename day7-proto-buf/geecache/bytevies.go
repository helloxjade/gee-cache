package geecache
//A ByteView holds an immutable view of bytes
type ByteView struct {
	b []byte   //存储真实的缓存值，选择byte为了能支持任意类型的存储  如字符串 图片
}
// Len returns the view's length
func(v ByteView) Len()int  {   //我们在 lru.Cache 的实现中，要求被缓存对象必须实现 Value 接口，即 Len() int 方法，返回其所占的内存大小
	return len(v.b)
}
// ByteSlice returns a copy of the data as a byte slice
func (v ByteView)ByteSlice()[]byte  {   //b 是只读的，使用 ByteSlice() 方法返回一个拷贝，防止缓存值被外部程序修改
	return cloneBytes(v.b)
}
// String returns the data as a string, making a copy if necessary
func(v ByteView) String()string  {
	return string(v.b)
}

func cloneBytes(b []byte)[]byte  {
	c:=make([]byte,len(b))
	copy(c,b)
	return c
}

