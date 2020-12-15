package main

import "ganCache/cacheCluster"

//也可以多开几个controller互相监视，其中一个为master controller，其他为slave controller。
//当master controller挂了，就第一个slave controller替换成新的master controller
func main()  {
	controller := cacheCluster.NewCacheClusterController()
	createCluster(controller)
	controller.CheckNodeStatus()
}

func createCluster(controller *cacheCluster.CacheClusterController)  {
	addrs := []string{
		"127.0.0.1:9001",
		"127.0.0.1:9002",
		"127.0.0.1:9003",
		"127.0.0.1:9004",
		"127.0.0.1:9005",
		"127.0.0.1:9006",
		"127.0.0.1:9007",
		"127.0.0.1:9008",
		"127.0.0.1:9009",
	}
	controller.CreateCluster(2, addrs)
}
