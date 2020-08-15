/*
搭建http服务端
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"main/Lincache"
	"net/http"
)

var db = map[string]string{
	"Tom":   "444",
	"Kate":  "589",
	"Linda": "312",
	"Sam":   "325",
}

func createGroup() *Lincache.Group {
	return Lincache.NewGroup("ID", 1<<11, Lincache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gro *Lincache.Group) {
	peers := Lincache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gro.RegisterPeers(peers)
	log.Println("lincache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gro *Lincache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gro.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fonted server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var (
		port int
		api  bool
	)
	flag.IntVar(&port, "port", 8001, "Lincache server port")
	flag.BoolVar(&api, "api", false, "Start api server?")
	flag.Parse()

	apiAddr := "http://localhost:8080"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], []string(addrs), group)
}
