package lraft

import (
	"fmt"
	"github.com/gangantongxue/ggl"
	"sync"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/hashicorp/raft"
)

// Guide 引导节点
func Guide() error {
	if err := global.Landlady.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(global.Opts.NodeID),
				Address: raft.ServerAddress(global.Opts.PeerAddr),
			},
		},
	}).Error(); err != nil {
		return err
	}
	if err := AddPeers(); err != nil {
		return err
	}
	return nil
}

type Signal struct {
	state bool
	lock  sync.Mutex
	cond  *sync.Cond
}

// AddPeers 添加节点
func AddPeers() error {
	if !global.NodeNum.IsFull() {
		signal := &Signal{}
		signal.cond = sync.NewCond(&signal.lock)
		go ListenPeerNode(signal)
		for !global.NodeNum.IsFull() {
		}
		for _, node := range global.NodesInfo.Get() {
			if node.ID == raft.ServerID(global.Opts.NodeID) {
				continue
			}
			if err := global.Landlady.AddVoter(node.ID, node.Address, 0, 0).Error(); err != nil {
				ggl.Error("add voter failed :", ggl.Err(err))
				signal.lock.Lock()
				signal.state = false
				signal.cond.Broadcast()
				signal.lock.Unlock()
				return err
			} else {
				ggl.Info(fmt.Sprintf("add voter(ID:%v,Address:%v) success", node.ID, node.Address))
			}
		}
		signal.lock.Lock()
		signal.state = true
		signal.cond.Broadcast()
		signal.lock.Unlock()
		ggl.Info("add peers success")
	}
	return nil
}
