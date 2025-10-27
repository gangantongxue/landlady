package lraft

import (
	"github.com/gangantongxue/ggl"
	"net"
	"strings"

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
	for !global.NodeNum.IsFull() {
		conn, err := ln.Accept()
		if err != nil {
			ggl.Error("accept peer node failed :", ggl.Err(err))
			continue
		}
		go func(conn net.Conn) {
			buf := make([]byte, 1024)
			_, err := conn.Read(buf)
			if err != nil {
				ggl.Error("read peer node info failed :", ggl.Err(err))
				return
			}
			global.NodesInfo.Add(model.Node{
				ID:      raft.ServerID(strings.Split(string(buf), ",")[0]),
				Address: raft.ServerAddress(strings.Split(string(buf), ",")[1]),
			})
			global.NodeNum.Inc()
			defer conn.Close()

			signal.lock.Lock()
			signal.cond.Wait()
			signal.lock.Unlock()

			if signal.state {
				conn.Write([]byte("success"))
			} else {
				conn.Write([]byte("failed"))
			}

		}(conn)
	}
}
