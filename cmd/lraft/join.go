package lraft

import (
	"github.com/gangantongxue/ggl"
	"net"

	"github.com/gangantongxue/landlady/cmd/global"
)

// Join 加入集群
func Join() error {
	conn, err := net.Dial("tcp", global.Opts.GuideAddr)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	defer conn.Close()
	conn.Write([]byte(global.Opts.NodeID + "," + global.Opts.PeerAddr))

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	if string(buf) != "success" {
		ggl.Error("join cluster failed")
		return err
	}
	nodeNum, err := global.NodesInfo.Flush(global.Landlady)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	global.NodeNum.SetCurNum(nodeNum)
	ggl.Info("join cluster success", ggl.Int("nodeNum", nodeNum))
	return nil
}
