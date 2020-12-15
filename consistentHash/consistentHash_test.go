package consistentHash

import (
	"reflect"
	"testing"
)

func TestMap_AddNodes(t *testing.T) {
	m := Default()
	err := m.AddNodes([]string{"192.168.1.101:9000", "192.168.1.102:9000", "192.168.1.103:9000"})
	if err != nil {
		t.Error(err)
	}
	hashs := make([]uint32, 0)
	for _, node := range m.Nodes{
		hashs = append(hashs, node.Hash)
	}
	var exceptHashs = []uint32{0, 25, 50}
	if !reflect.DeepEqual(hashs, exceptHashs) {
		t.Errorf("except hashs : \n %v \n but \n %v", exceptHashs, hashs)
	}

	err = m.AddNode("192.168.1.104:9000")
	if err != nil {
		t.Error(err)
	}
	if m.Nodes[3].Hash != 75 {
		t.Errorf("except hash : 75 , but  %v", m.Nodes[3].Hash)
	}
}

func TestMap_SearchNode(t *testing.T) {
	m := Default()
	err := m.AddNodes([]string{"192.168.1.101:9000", "192.168.1.102:9000", "192.168.1.103:9000"})
	if err != nil {
		t.Error(err)
	}
	nodes := m.Nodes

	key1 := []byte{0}
	node1 := m.SearchNode(key1)
	if node1 != nodes[0].Id {
		t.Errorf("TestMap_SearchNode \n except : %s \n , but  %s", nodes[0].Id, node1)
	}

	key2 := []byte{35}
	node2 := m.SearchNode(key2)
	if node2 != nodes[2].Id {
		t.Errorf("TestMap_SearchNode \n except : %s \n , but  %s", nodes[2].Id, node2)
	}

	key3 := []byte{55}
	node3 := m.SearchNode(key3)
	if node3 != nodes[0].Id {
		t.Errorf("TestMap_SearchNode \n except node1 : %s \n , but  %s", nodes[0].Id, node3)
	}

	key4 := []byte{10, 5}
	node4 := m.SearchNode(key4)
	if node4 != nodes[1].Id {
		t.Errorf("TestMap_SearchNode \n except node1 : %s \n , but  %s", nodes[1].Id, node4)
	}

	m.AddNode("192.168.1.104:9000")
	nodes = m.Nodes
	key5 := []byte{55}
	node5 := m.SearchNode(key5)
	if node5 != nodes[3].Id {
		t.Errorf("TestMap_SearchNode \n except  : %s \n , but  %s", nodes[3].Id, node5)
	}
}

