package cacheCluster

import (
	"ganCache/consistentHash"
)

//在consistentHash.HashMap基础上封装一层，增添一些功能
type CacheClusterHashMap struct {
	//hash分布
	HM *consistentHash.HashBelongMap
	//一个Group包含一个master和多个slave
	Groups []*MasterAndSlave
}

type MasterAndSlave struct {
	//对应*consistentHash.HashBelongMap的node的id
	Id string
	//ip:port
	Master string
	//ip:port0
	Slaves []string
}

type NodeMapInfo struct {
	Peers []*consistentHash.Peer
	Groups []*MasterAndSlave
}

func NewCacheClusterHashMap() *CacheClusterHashMap {
	cc := CacheClusterHashMap{}
	cc.HM = consistentHash.Default()
	cc.Groups = make([]*MasterAndSlave, 0)
	return &cc
}


func (cc *CacheClusterHashMap) SearchGroup(key string) *MasterAndSlave {
	nodeId := cc.HM.SearchPeer(key)
	for _, group := range cc.Groups {
		if group.Id == nodeId {
			return group
		}
	}
	//按理说，任何key都分配在了hash圈上，如果没找到就是未知错误，panic
	panic(key + "can not search node")
}

func (cc *CacheClusterHashMap) UpdateNodeMapInfo(info *NodeMapInfo) {
	cc.HM.Peers = info.Peers
	cc.Groups = info.Groups
}


