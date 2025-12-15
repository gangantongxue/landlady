package lraft

import (
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
		// 移除无限循环，让引导节点先完成初始化
		// 新节点加入将在ListenPeerNode中异步处理
	}
	return nil
}
