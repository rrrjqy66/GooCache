package goocache

// ByteView是一个只读的数据结构，底层是一个字节切片。
// 它提供了Len方法来获取数据的长度，以及ByteSlice方法来获取数据的副本。
// 由于ByteView是不可变的，所以它可以安全地在多个goroutine之间共享，而不需要担心数据竞争问题。
// 这使得ByteView成为缓存系统中非常有用的数据结构，可以高效地存储和传递数据，同时保证线程安全。
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) //返回一个副本，防止外部修改底层数据
}
func (v ByteView) String() string {
	return string(v.b) //string底层也是一个字节切片直接转换
}
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c

}
