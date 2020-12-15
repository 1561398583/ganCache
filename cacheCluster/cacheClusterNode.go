package cacheCluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ganCache/cache2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)


//ClusterNode负责对数据的读写

type ClusterNode struct {
	//ip:port
	SelfAddr string
	//节点hash分布
	CCHM *CacheClusterHashMap
	//本地缓存
	cache *cache2.SafeCache
}

func NewClusterNode(addr string) *ClusterNode {
	node := ClusterNode{}
	node.SelfAddr = addr
	node.CCHM = NewCacheClusterHashMap()
	node.cache = cache2.NewSafeCache(100 * 1024 * 1024)	//100M
	return &node
}

func (p *ClusterNode) ServeHTTP(w http.ResponseWriter, r *http.Request){
	path := strings.Trim(r.URL.Path, "/") //裁剪path两边的‘/’.
	pathParts := strings.Split(path, "/")
	handl := pathParts[0]
	fmt.Println("handle : " + handl)
	switch handl {
	case "get":
		p.Get(w, r, pathParts)
	case "set":
		p.Set(w, r, pathParts)
	case "isalive":
		p.IsAlive(w, r)
	case "nodemap":
		p.NodeMap(w, r)
	case "updatenodemapinfo":
		p.updateNodeMapInfo(w, r)
	default:
		http.Error(w, "method error", http.StatusBadRequest)
		return
	}
}

//key是否存在于本机
func (p *ClusterNode) IsMe(key string, masterAndSlave *MasterAndSlave) bool{
	if masterAndSlave.Master == p.SelfAddr {
		return true
	}
	for _, slave := range masterAndSlave.Slaves {
		if slave == p.SelfAddr {
			return true
		}
	}
	return false
}

func (p *ClusterNode) Set(w http.ResponseWriter, r *http.Request, pathParts []string){
	key := pathParts[1]
	log.Println("set " + key)
	masterAndSlave := p.CCHM.SearchGroup(key)
	if !p.IsMe(key, masterAndSlave) {
		//key不在我这，去其他节点看看
		log.Println("not belong to me")
		w.WriteHeader(http.StatusSeeOther)
		otherAddrs := masterAndSlave.Master
		for _, slave := range masterAndSlave.Slaves {
			otherAddrs += " "
			otherAddrs += slave
		}
		w.Write([]byte(otherAddrs))
		return
	}

	contentType := r.Header.Get("Content-Type")
	contentLenth := r.Header.Get("Content-Length")
	log.Println("contentType : " + contentType + "; lenth : " + contentLenth)
	fmt.Println()

	//value在body中
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	/*
		contentSize, _ := strconv.Atoi(contentLenth)
		bs := make([]byte, contentSize)
		r.Body.Read(bs)	//一次只能读取一帧的数据，要读取完整数据必须循环read，直到遇到EOF
	*/
	//取出完整的body数据
	value, _ := ioutil.ReadAll(r.Body)

	//加入cache
	p.cache.ADD(key, value)

	//如果自己是master，那么就同步数据给自己的slave
	if masterAndSlave.Master == p.SelfAddr {
		log.Println("i am master, so i must copy data to slaves")
		for _, slave := range masterAndSlave.Slaves {
			url := "http://" + slave + "/set/" + key
			resp, err := http.Post(url, contentType, bytes.NewReader(value))
			if err != nil {
				log.Println(err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Println(resp.Status)
			}
			err = resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}

	return
}

func (p *ClusterNode) Get(w http.ResponseWriter, r *http.Request, pathParts []string){
	key := pathParts[1]
	log.Println("get " + key)
	masterAndSlave := p.CCHM.SearchGroup(key)
	if !p.IsMe(key, masterAndSlave) {
		//key不属于我，去其他节点看看
		log.Println("not belong to me")
		w.WriteHeader(http.StatusSeeOther)
		otherAddrs := masterAndSlave.Master
		for _, slave := range masterAndSlave.Slaves {
			otherAddrs += " "
			otherAddrs += slave
		}
		w.Write([]byte(otherAddrs))
		return
	}

	value, ok := p.cache.Get(key)
	if !ok {	//key虽然属于我，但是还没存进来
		log.Println("belong to me but not found")
		http.Error(w, "", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err := w.Write(value)	//Write会自动添加header（http.StatusOk, Content-Lenth）
	if err != nil {
		fmt.Println(err)
		return
	}
}

/*
//创建集群
//body的格式(以空格分割)：
//slaveNumber ip1:port1 ip2:port2 ...
func (p *ClusterNode) CreateCluster(w http.ResponseWriter, r *http.Request){
	//已经创建了集群
	if len(p.CCHM.Groups) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cluster have already create"))
		return
	}
	//取出完整的body数据
	value, _ := ioutil.ReadAll(r.Body)
	//转为string
	vs := string(value)
	fmt.Println("createCluster get :" + vs)
	//以空格分割
	parts := strings.Split(vs, " ")

	slaveNum ,err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "first string must be number", http.StatusBadRequest)
		return
	}
	//加入hashMap
	err = p.CCHM.AddNodes(slaveNum, parts[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("create cluster success")
	for n, node := range p.CCHM.Groups {
		fmt.Println("id=" + p.CCHM.HM.Peers[n].Id +" ; hashValue=" + strconv.FormatInt(int64(p.CCHM.HM.Peers[n].HashValue), 10))
		fmt.Println("master : " + node.Master)
		slaveMsg := "slaves : "
		for _,slave := range node.Slaves {
			slaveMsg += slave
			slaveMsg += "   "
		}
		fmt.Println(slaveMsg)
		fmt.Println()
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
	return
}


func (p *ClusterNode) AddNodes(w http.ResponseWriter, r *http.Request){
	//取出完整的body数据
	value, _ := ioutil.ReadAll(r.Body)
	//转为string
	vs := string(value)
	fmt.Println("AddNodes : " + vs)
	//以空格分割
	parts := strings.Split(vs, " ")

	slaveNum ,err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "first string must be number", http.StatusBadRequest)
		return
	}
	//加入hashMap
	err = p.CCHM.AddNodes(slaveNum, parts[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("AddNodes success, Now map is :")
	for n, node := range p.CCHM.Groups {
		fmt.Println("id=" + p.CCHM.HM.Peers[n].Id +" ; hashValue=" + strconv.FormatInt(int64(p.CCHM.HM.Peers[n].HashValue), 10))
		fmt.Println("master : " + node.Master)
		slaveMsg := "slaves : "
		for _,slave := range node.Slaves {
			slaveMsg += slave
			slaveMsg += "   "
		}
		fmt.Println(slaveMsg)
		fmt.Println()
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
	return
}
 */

//用于client检测server是否掉线
func (p *ClusterNode) IsAlive(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
	return
}

//NodeHash分布信息
func (p *ClusterNode) NodeMap(w http.ResponseWriter, r *http.Request) {
	nodeMapInfo := NodeMapInfo{}
	nodeMapInfo.Peers = p.CCHM.HM.Peers
	nodeMapInfo.Groups = p.CCHM.Groups
	bs, err := json.Marshal(nodeMapInfo)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("nodeMapInfo : ")
	fmt.Println(string(bs))
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
	return
}

//把nodeMap复制到本节点
func (p *ClusterNode) updateNodeMapInfo(w http.ResponseWriter, r *http.Request){
	//取出完整的body数据
	value, _ := ioutil.ReadAll(r.Body)
	nodeMapInfo := NodeMapInfo{}
	//json转为对象
	err := json.Unmarshal(value, &nodeMapInfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	p.CCHM.UpdateNodeMapInfo(&nodeMapInfo)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
	return
}







