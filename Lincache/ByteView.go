package Lincache

type ByteView struct {
	val []byte // 使用byte数组可以表示多种数据，eg. 字符串，图片
}

func (v ByteView) Len() int {
	// 实现Len接口
	return len(v.val)
}

func (v ByteView) ByteSlice() []byte {
	// 返回一份当前数据的拷贝
	tmp := make([]byte, len(v.val))
	copy(tmp, v.val)
	return tmp
}

func (v ByteView) String() string {
	// byte数组转换成string
	return string(v.val)
}
