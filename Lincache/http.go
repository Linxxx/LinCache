package Lincache

import (
	"fmt"
	"io/ioutil"
	"log"
	"main/Lincache/consistenthash"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/* Http服务端 */
type HTTPPool struct {
	self, basePath string
	mu             sync.Mutex             // 互斥锁
	peers          *consistenthash.Map    // 根据key选择节点
	httpGetters    map[string]*httpGetter // 记录每个远程节点和客户端的映射关系
}

const (
	defaultBasePath = "/_lincache/"
	defaultReplicas = 50
)

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		// self是自身url，basePath是其他节点的基址
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 判断
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (p *HTTPPool) Set(peers ...string) {
	// 设置同级节点
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath} // 为每个节点添加客户端
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	// 获取需要访问的节点的客户端
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

/* Http客户端 */
type httpGetter struct {
	baseURL string // 表示将要访问的远程节点的地址
}

func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // 关闭响应体

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body) // 读取响应体的数据
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
var _ PeerPicker = (*HTTPPool)(nil)
