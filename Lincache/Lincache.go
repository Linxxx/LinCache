/*
支持并发访问和不同用户组的LRU缓存，数据格式为string: []byte
*/
package Lincache

import (
	"fmt"
	"log"
	"main/Lincache/Lincachepb"
	"main/Lincache/singleflight"
	"sync"
)

type Getter interface {
	// 该接口用于当缓存未命中时获取源数据的功能
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	// 缓存命名空间
	name     string
	getter   Getter
	lincache cache
	peers    PeerPicker
	loader   *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxmemosize int64, getter Getter) *Group {
	// group的构造函数
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:     name,
		getter:   getter,
		lincache: cache{maxmemosize: maxmemosize},
		loader:   &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.lincache.get(key); ok {
		log.Println("[Lincache] Hit")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		return g.loadFromPeer(key)
	})

	if err == nil {
		return view.(ByteView), err
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{val: bytes} // 深拷贝？ *****
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.lincache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) loadFromPeer(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if val, err := g.getFromPeer(peer, key); err == nil {
				return val, nil
			}
			log.Println("[LinCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &Lincachepb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &Lincachepb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{res.Value}, nil
}
