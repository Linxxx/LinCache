package Lincache

import (
	"main/Lincache/Lincachepb"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// rpc调用方式
	Get(in *Lincachepb.Request, out *Lincachepb.Response) error
}
