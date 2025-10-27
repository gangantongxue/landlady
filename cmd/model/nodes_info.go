package model

import (
	"github.com/gangantongxue/ggl"
	"sync"

	"github.com/hashicorp/raft"
)

type NodesInfo struct {
	m     sync.RWMutex
	nodes []Node
}

type Node raft.Server

// Flush 刷新节点
func (n *NodesInfo) Flush(l *raft.Raft) (int, error) {
	cfgFuture := l.GetConfiguration()
	if err := cfgFuture.Error(); err != nil {
		ggl.Error("flush nodes info error :", ggl.Err(err))
		return 0, err
	}
	cfg := cfgFuture.Configuration()
	n.m.Lock()
	defer n.m.Unlock()
	n.nodes = []Node{}
	for _, server := range cfg.Servers {
		n.nodes = append(n.nodes, Node(server))
	}
	return len(n.nodes), nil
}

// Add 添加节点
func (n *NodesInfo) Add(node Node) {
	n.m.Lock()
	defer n.m.Unlock()
	n.nodes = append(n.nodes, node)
}

// Get 获取节点
func (n *NodesInfo) Get() []Node {
	n.m.RLock()
	defer n.m.RUnlock()
	return n.nodes
}

// Del 删除节点
func (n *NodesInfo) Del(id raft.ServerID) {
	n.m.Lock()
	defer n.m.Unlock()
	for i, node := range n.nodes {
		if node.ID == id {
			n.nodes = append(n.nodes[:i], n.nodes[i+1:]...)
			break
		}
	}
}
