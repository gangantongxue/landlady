package lraft

import (
	"errors"
	"github.com/gangantongxue/ggl"
	"net"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
)

// Join 加入集群
func Join() error {
	// 设置连接超时
	conn, err := net.DialTimeout("tcp", global.Opts.GuideAddr, 5*time.Second)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	defer conn.Close()
	
	// 发送节点信息
	_, err = conn.Write([]byte(global.Opts.NodeID + "," + global.Opts.PeerAddr))
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	
	// 正确检查响应结果，只比较实际读取的字节
	response := string(buf[:n])
	if response != "success" {
		ggl.Error("join cluster failed, response:", ggl.Str("response", response))
		return errors.New("join cluster failed: " + response)
	}
	
	// 刷新节点信息
	nodeNum, err := global.NodesInfo.Flush(global.Landlady)
	if err != nil {
		ggl.Error("join cluster failed :", ggl.Err(err))
		return err
	}
	
	// 设置当前节点数量
	global.NodeNum.SetCurNum(nodeNum)
	ggl.Info("join cluster success", ggl.Int("nodeNum", nodeNum))
	return nil
}
