package cacheHttp

import (
	"log"
	"net/http"
	"testing"
)

func TestCacheHttp_ServeHTTP(t *testing.T) {
	server := NewCacheHttp(1000 * 1024 * 1024)
	log.Fatal(http.ListenAndServe(":9000", server))
}
