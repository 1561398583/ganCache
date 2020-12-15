package cacheCluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ganCache/consistentHash"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

//CacheClusterController负责对CacheClusterNode的控制

type CacheClusterController struct {
	//节点hash分布
	CCHM *CacheClusterHashMap
	//所有节点ip:port
	Nodes []string
	//是否已经建立集群
	IsCluster	bool
	//下一个group的id
	NewGroupId int
}

func NewCacheClusterController() *CacheClusterController{
	hashMap := NewCacheClusterHashMap()
	controller := CacheClusterController{}
	controller.CCHM = hashMap
	nodes := make([]string, 0)
	controller.Nodes = nodes
	return &controller
}

func (cc *CacheClusterController) getNewId() string{
	cc.NewGroupId ++
	return strconv.FormatInt(int64(cc.NewGroupId), 10)
}


//每一个group有一个master和slaveNum个slave
func (controller *CacheClusterController) CreateCluster(slaveNum int, nodes []string) {
	//测试所有节点是否能连接
	notAliveNodes := isAllAlive(nodes)
	if len(notAliveNodes) != 0 {
		fmt.Println("CreateCluster error , these node can not connect")
		for _, node := range notAliveNodes {
			fmt.Println(node)
		}
		return
	}

	//所有节点都能正常链接，于是开始创建集群
	//先构建hash分布map
	groups, err := controller.createGroups(slaveNum, nodes)
	if err != nil {
		fmt.Println("CreateCluster error : " + err.Error())
		return
	}

	controller.CCHM = NewCacheClusterHashMap()
	controller.CCHM.Groups = groups
	//获取groups的所有id
	ids := make([]string, 0)
	for _, g := range groups {
		ids = append(ids, g.Id)
	}
	err = controller.CCHM.HM.AddPeers(ids)
	if err != nil {
		fmt.Println("CreateCluster error : " + err.Error())
		return
	}

	//加入Nodes
	for _, node := range nodes {
		controller.Nodes = append(controller.Nodes, node)
	}

	//然后更新每个node的nodeMapInfo
	err = controller.updateNodeMapInfoToAllNodes()
	if err != nil {
		fmt.Println("CreateCluster error : " + err.Error())
		return
	}

	controller.IsCluster = true	//已经构建集群

	fmt.Println("create cluster sucess, now node map :")
	fmt.Println(string(controller.getNodeMapInfo()))
}


//slaveNum： slave数量
//newAddrs：新添加的节点地址
func (controller *CacheClusterController) AddNodes(slaveNum int, newAddrs []string)  {
	//是否已经构建集群
	if !controller.IsCluster {
		fmt.Printf("please create cluster first")
		return
	}

	//测试所有节点是否能连接
	notAliveNodes := isAllAlive(newAddrs)
	if len(notAliveNodes) != 0 {
		fmt.Println("AddNodes error , these node can not connect")
		for _, node := range notAliveNodes {
			fmt.Println(node)
		}
		return
	}

	//都可以连接，继续
	//构建新的group
	groups, err := controller.createGroups(slaveNum, newAddrs)
	if err != nil {
		fmt.Println("AddNodes error : " + err.Error())
		return
	}
	//加入cluster groups
	for _, group := range groups{
		controller.CCHM.Groups = append(controller.CCHM.Groups, group)
	}
	//加入hash分布map
	//获取groups的所有id
	ids := make([]string, 0)
	for _, g := range groups {
		ids = append(ids, g.Id)
	}
	err = controller.CCHM.HM.AddPeers(ids)
	if err != nil {
		fmt.Println("CreateCluster error : " + err.Error())
		return
	}

	//加入新节点
	for _, addr := range newAddrs {
		controller.Nodes = append(controller.Nodes, addr)
	}

	//然后更新每个node的nodeMapInfo
	err = controller.updateNodeMapInfoToAllNodes()
	if err != nil {
		fmt.Println("CreateCluster error : " + err.Error())
		return
	}

	fmt.Println("addNodes sucess, now node map :")
	fmt.Println(string(controller.getNodeMapInfo()))
}




//每一个group有一个master和slaveNum个slave
func (controller *CacheClusterController) createGroups(slaveNum int, nodes []string) ([]*MasterAndSlave,error) {
	//必须是主备数量的整数倍
	if len(nodes) % (slaveNum + 1) != 0 {
		return nil, errors.New("addr number error")
	}
	groups := make([]*MasterAndSlave, 0)
	nodeIndex := 0
	for {
		mas := MasterAndSlave{}
		mas.Id = controller.getNewId()
		mas.Master = nodes[nodeIndex]
		nodeIndex ++
		mas.Slaves = make([]string, slaveNum)
		for i := 0; i < slaveNum; i++ {
			mas.Slaves[i] = nodes[nodeIndex]
			nodeIndex ++
		}
		groups = append(groups, &mas)

		if nodeIndex == len(nodes) {
			break
		}
	}

	return groups, nil
}


//测试节点是否能正常连接
func isAlive(addr string) bool {
	alive := true
	url := "http://" + addr + "/isalive"
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		alive = false
	}
	if resp != nil {
		err := resp.Body.Close()
		if err != nil {
			//预期外的error，打印栈信息
			errWithStack := errors.New(err.Error())
			fmt.Printf("%+v", errWithStack)
		}
	}
	return alive
}

//测试所有节点是否能正常连接
func isAllAlive(addrs []string) []string {
	notAlive := make([]string, 0)
	for _, addr := range addrs {
		alive := isAlive(addr)
		if !alive {
			notAlive = append(notAlive, addr)
		}
	}
	return notAlive
}


//监控每个节点的状态
func (controller *CacheClusterController) CheckNodeStatus() {
	for {
		time.Sleep(time.Minute)

		//询问每个node是否正常
		deadNodes := isAllAlive(controller.Nodes)


		//踢出deadNode
		if len(deadNodes) > 0 {
			fmt.Println("these node have dead : ")
			for _, dn := range deadNodes {
				fmt.Println(dn)
			}

			controller.takeOutDeadNode(deadNodes)
			//更新所有node的mapInfo
			err := controller.updateNodeMapInfoToAllNodes()
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

//去掉掉线的node
func (controller *CacheClusterController) takeOutDeadNode(deadNodes []string) {
	//查看每个MasterAndSlave
	groups := controller.CCHM.Groups
	for _, group := range groups {
		err := createNewMasterAndSlave(group, deadNodes)
		if err != nil {
			//说明这个group的全部node都死亡，于是去掉这个group
			fmt.Println(group.Id + " " + err.Error())
			controller.takeOutDeadGroup(group.Id)
		}
	}

	//从Nodes中去掉
	newNodes := make([]string, 0)
	for _, node := range controller.Nodes {
		if !isIn(node, deadNodes) {	//不在死亡名单中
			newNodes = append(newNodes, node)
		}
	}
	controller.Nodes = newNodes
}

//去掉掉线的group
func (controller *CacheClusterController) takeOutDeadGroup(groupId string) {
	//更新peers
	newPeers := make([]*consistentHash.Peer, 0)
	for _, peer := range controller.CCHM.HM.Peers {
		if peer.Id != groupId {
			newPeers = append(newPeers, peer)
		}
	}
	controller.CCHM.HM.Peers = newPeers

	//更新groups
	newGroups := make([]*MasterAndSlave, 0)
	for _, group := range controller.CCHM.Groups {
		if group.Id != groupId {
			newGroups = append(newGroups, group)
		}
	}
	controller.CCHM.Groups = newGroups
}

//更新所有node的nodeMapinfo
func (controller *CacheClusterController) updateNodeMapInfoToAllNodes() error{
	bs := controller.getNodeMapInfo()
	fmt.Println("updateNodeMapInfo : ")
	fmt.Println(string(bs))

	//把nodeMapInfo复制到所有节点中
	for _, node := range controller.Nodes{
		url := "http://" + node + "/updatenodemapinfo"
		sendData := bytes.NewReader(bs)
		resp, err := http.Post(url, "text/plain", sendData)
		if err != nil {
			return errors.New("updateNodeMapInfo error : " + err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintln("updateNodeMapInfo error")
			msg += fmt.Sprintln(node + " resp status : " + resp.Status)
			return errors.New(msg)
		}
		resp.Body.Close()
	}
	return nil
}

func (controller *CacheClusterController) getNodeMapInfo() []byte {
	nodeMapInfo := NodeMapInfo{}
	nodeMapInfo.Peers = controller.CCHM.HM.Peers
	nodeMapInfo.Groups = controller.CCHM.Groups
	bs, err := json.Marshal(nodeMapInfo)
	if err != nil {
		panic(err)
	}
	return bs
}


//b是否包含a
func isIn(a string, b []string) bool {
	for _, elem := range b {
		if a == elem {
			return true
		}
	}
	return false
}

func createNewMasterAndSlave(mas *MasterAndSlave, deadNode []string) error {
	newSlaves := make([]string, 0)
	//查看是否有slave死亡
	for _, slave := range mas.Slaves{
		//如果没死，就加入newSlaves
		if !isIn(slave, deadNode) {
			newSlaves = append(newSlaves, slave)
		}
	}
	mas.Slaves = newSlaves

	//如果master在死亡名单上,那么就用第一个slave来替换master
	if isIn(mas.Master, deadNode) {
		//slave也都死了，于是这个group就没可用的节点了
		if len(mas.Slaves) == 0 {
			return errors.New("this group node all die")
		}else {
			mas.Master = mas.Slaves[0]
			mas.Slaves = mas.Slaves[1:]
		}
	}

	return nil
}


