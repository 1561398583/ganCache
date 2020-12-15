package cacheHttp

import (
	"fmt"
	"ganCache/cache2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//提供http功能
//url格式：
//"https://example.net:8000/<key>"

type CacheHttp struct {
	cache *cache2.SafeCache
}

//self ip:port
func NewCacheHttp(cacheSize int64) *CacheHttp {
	c := cache2.NewSafeCache(cacheSize)
	return &CacheHttp{
		cache: c,
	}
}

func getter(key string) ([]byte, error) {
	return nil, nil
}

func (p *CacheHttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/") //裁剪path两边的‘/’.
	pathParts := strings.Split(path, "/")
	handl := pathParts[0]
	switch handl {
	case "get":
		p.Get(w, r, pathParts)
	case "set":
		p.Set(w, r, pathParts)
	default:
		http.Error(w, "method error", http.StatusBadRequest)
		return
	}
}

func (p *CacheHttp) Set(w http.ResponseWriter, r *http.Request, pathParts []string){
	key := pathParts[1]
	contentType := r.Header.Get("Content-Type")
	contentLenth := r.Header.Get("Content-Length")
	log.Println("set " + key)
	log.Println("contentType : " + contentType + "; lenth : " + contentLenth)
	fmt.Println()

	//value在body中
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			panic(err)
		}
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			panic(err)
		}
		return
	}

	/*
	contentSize, _ := strconv.Atoi(contentLenth)
	bs := make([]byte, contentSize)
	r.Body.Read(bs)	//一次只能读取一帧的数据，要读取完整数据必须循环read，直到遇到EOF
	*/


	value, _ := ioutil.ReadAll(r.Body)

	//加入cache
	p.cache.ADD(key, value)

	return
}

func (p *CacheHttp) Get(w http.ResponseWriter, r *http.Request, pathParts []string){
	key := pathParts[1]
	log.Println("get " + key)

	value, ok := p.cache.Get(key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(""))
		if err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err := w.Write(value)	//Write会自动添加header（http.StatusOk, Content-Lenth）
	if err != nil {
		fmt.Println(err)
		return
	}
}

