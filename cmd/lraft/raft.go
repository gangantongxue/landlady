package lraft

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/model"
	"github.com/gangantongxue/landlady/cmd/opts"
	"github.com/hashicorp/raft"
)

// Init 初始化
func Init() error {
	global.NodeNum.SetPreNum(global.Opts.NodeNum)
	global.NodesInfo.Add(model.Node{
		ID:      raft.ServerID(global.Opts.NodeID),
		Address: raft.ServerAddress(global.Opts.PeerAddr),
	})
	global.NodeNum.Inc()

	var err error
	if global.Landlady, err = Start(global.Opts); err != nil {
		return err
	}
	if global.Opts.Join {
		if err := Join(); err != nil {
			return err
		}
	} else {
		if err := Guide(); err != nil {
			return err
		}
	}
	return nil
}

// Start 开启节点
func Start(o *opts.Opts) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(o.NodeID)
	config.HeartbeatTimeout = 500 * time.Millisecond
	config.ElectionTimeout = 500 * time.Millisecond
	config.LeaderLeaseTimeout = 500 * time.Millisecond
	config.CommitTimeout = 50 * time.Millisecond

	raftDataDir := filepath.Join(o.DataDir, "landlady_raft")

	// 确保raftDataDir目录存在
	if err := os.MkdirAll(raftDataDir, 0755); err != nil {
		return nil, err
	}

	// 创建文件快照存储
	snapshots, err := raft.NewFileSnapshotStore(raftDataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	// 创建TCP transport
	addr, err := net.ResolveTCPAddr("tcp", o.PeerAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(o.PeerAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// 使用内存存储，简化测试
	return raft.NewRaft(config, global.FSM, raft.NewInmemStore(), raft.NewInmemStore(), snapshots, transport)
}
