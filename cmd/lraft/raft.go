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
	raftboltdb "github.com/hashicorp/raft-boltdb"
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
		Join()
	} else {
		Guide()
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

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDataDir, "landlady_log.bolt"))
	if err != nil {
		return nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDataDir, "landlady_stable.bolt"))
	if err != nil {
		return nil, err
	}
	snapshots, err := raft.NewFileSnapshotStore(raftDataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveTCPAddr("tcp", o.PeerAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(o.PeerAddr, addr, o.NodeNum, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	raftNode, err := raft.NewRaft(config, global.FSM, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, err
	}

	return raftNode, nil
}
