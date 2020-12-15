package main

import (
	"fmt"
	"ganCache/cacheCluster"
	"log"
	"net/http"
)

func main()  {
	addr := "127.0.0.1:9006"
	node := cacheCluster.NewClusterNode(addr)
	fmt.Println(addr + " begin")
	log.Fatal(http.ListenAndServe(addr, node))
}
