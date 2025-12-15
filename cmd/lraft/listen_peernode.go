package lraft

import (
	"fmt"
	"net"
	"strings"

	"github.com/gangantongxue/ggl"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/model"
	"github.com/hashicorp/raft"
)

func ListenPeerNode(signal *Signal) {
	ln, err := net.Listen("tcp", ":"+global.Opts.GuidePort)
	if err != nil {
		ggl.Error("listen peer node failed :", ggl.Err(err))
		return
	}

	// 不依赖global.NodeNum.IsFull()，持续监听节点加入请求
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				ggl.Error("accept peer node failed :", ggl.Err(err))
				continue
			}
			go handlePeerNode(conn, signal)
		}
	}()
}

// handlePeerNode 处理单个节点加入请求
func handlePeerNode(conn net.Conn, signal *Signal) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		ggl.Error("read peer node info failed :", ggl.Err(err))
		conn.Write([]byte("failed"))
		return
	}

	// 解析节点信息
	parts := strings.Split(string(buf), ",")
	if len(parts) != 2 {
		ggl.Error("invalid peer node info")
		conn.Write([]byte("failed"))
		return
	}

	nodeID := raft.ServerID(parts[0])
	address := raft.ServerAddress(parts[1])

	// 添加节点到全局信息
	global.NodesInfo.Add(model.Node{
		ID:      nodeID,
		Address: address,
	})
	global.NodeNum.Inc()

	// 立即添加为投票节点
	if err := global.Landlady.AddVoter(nodeID, address, 0, 0).Error(); err != nil {
		ggl.Error("add voter failed :", ggl.Err(err))
		conn.Write([]byte("failed"))
		return
	}

	ggl.Info(fmt.Sprintf("add voter(ID:%v,Address:%v) success", nodeID, address))
	conn.Write([]byte("success"))
}
