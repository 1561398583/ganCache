package consistentHash

import (
	"github.com/pkg/errors"
	"sort"
)

type HashFunc func(key []byte) uint32

type Peer struct {
	Id string
	//Peer占据的hash值
	HashValue uint32
}

type HashBelongMap struct {
	//计算hash的func
	getHash HashFunc
	//hash值的个数，默认为100个
	hashSize uint32
	//节点
	Peers []*Peer
}

func (m *HashBelongMap) AddPeers(ids []string) error {
	for _, id := range ids {
		err := m.AddPeer(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *HashBelongMap) AddPeer(Id string) error{
	if len(m.Peers) == 0 {
		newPeer := Peer{Id: Id, HashValue: 0}
		m.Peers = append(m.Peers, &newPeer)
		return nil
	}

	//检查是否id重复
	for _, peer := range m.Peers {
		if peer.Id == Id {
			return errors.New("id has existed")
		}
	}

	//在Peers最后加一个哨兵Peer，解决首位连接处的问题（就不用在首位连接处做特殊判断并处理）
	//Id为""，
	//Hash值为hashSize（在hash值圈上并没有这个hash值，这是hash值圈上最后一个hash值加1，相当于在hash值圈上，100和0（起点）重合）
	guard := Peer{Id: "", HashValue: m.hashSize}
	//寻找一个间隔最大的
	//为了不改变原来的Peers，这里复制一份
	peersTemp := m.Peers[:]
	peersTemp = append(peersTemp, &guard)
	var maxInterval uint32 = 0
	maxIndex := 0
	//peer之间的间隔比peer数量少1
	for i := 0; i < len(peersTemp) - 1; i++ {
		interval := peersTemp[i+1].HashValue - peersTemp[i].HashValue
		if interval > maxInterval {
			maxInterval = interval
			maxIndex = i
		}
	}
	//找到间隔最大的node的索引和最大间隔值，然后计算新node的hash值
	newPeerHashValue := peersTemp[maxIndex].HashValue + (maxInterval / 2)
	//构造新node
	newPeer := Peer{Id: Id, HashValue: newPeerHashValue}
	//添加到原Peers
	m.Peers = append(m.Peers, &newPeer)
	//Nodes根据hash值排序
	sort.Sort(m)
	return nil
}

func (m *HashBelongMap) SearchPeer(key string) string {
	hash := m.getHash([]byte(key))
	index := 0
	for n, peer := range m.Peers {
		if peer.HashValue >= hash {
			//找到符合要求的peer，index赋值为该peer的索引；
			//如果循环完了也没找到，那index还是0，意思是就是第一个peer
			index = n
			break
		}
	}
	return m.Peers[index].Id
}

//实现排序的3个方法
func (m *HashBelongMap) Len() int{
	return len(m.Peers)
}

func (m *HashBelongMap) Less(i, j int) bool{
	if m.Peers[i].HashValue < m.Peers[j].HashValue {
		return true
	}
	return false
}

func (m *HashBelongMap) Swap(i, j int){
	node := m.Peers[i]
	m.Peers[i] = m.Peers[j]
	m.Peers[j] = node
}

func Default() *HashBelongMap{
	return New(100, defaultHash)
}

func New(hashSize uint32, getHash HashFunc)  *HashBelongMap{
	m := HashBelongMap{hashSize: hashSize, getHash: getHash, Peers: make([]*Peer, 0)}
	return &m
}

//默认hashFunc
//把所有字节的值相加，再对100取余数。总共有100个hash值
func defaultHash(key []byte) uint32  {
	var sum uint32 = 0
	for _, b := range key {
		sum += uint32(b)
	}
	return sum % 100
}
