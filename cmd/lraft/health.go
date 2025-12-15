package lraft

import (
	"net"
	"time"

	"github.com/gangantongxue/ggl"

	"github.com/gangantongxue/landlady/cmd/global"
)

// InitHealthCheck 初始化健康检查
func InitHealthCheck() {
	go func() {
		isLeaderCh := global.Landlady.LeaderCh()
		closeHealthCheck := make(chan struct{})
		for isLeader := range isLeaderCh {
			if isLeader {
				ggl.Info("become leader")
				startHealthCheck(closeHealthCheck)
			} else {
				close(closeHealthCheck)
			}
		}
	}()
}

// startHealthCheck 启动健康检查
func startHealthCheck(closeHealthCheck chan struct{}) {
	healthStatus := make(chan uint8)
	// 启动定时健康检查
	ticker := time.NewTicker(time.Duration(global.Opts.HealthCheckInterval) * time.Second)
	defer ticker.Stop()
	// 首次健康检查
	healthStatus <- HealthCheck()
	// 定时健康检查
	go func() {
		for {
			select {
			case <-closeHealthCheck:
				return
			case <-ticker.C:
				healthStatus <- HealthCheck()
			}
		}
	}()
	for {
		select {
		case <-closeHealthCheck:
			return
		case state := <-healthStatus:
			switch state {
			case 0:
			case 1:
				if err := AddPeers(); err != nil {
					ggl.Error("add peers failed :", ggl.Err(err))
				}
			case 2:
				ggl.Error("Cluster is unavailable")
				return
			}
		}
	}
}

// HealthCheck 健康检查
// 返回集群状态，0 健康，1 不健康，2 不可用
func HealthCheck() uint8 {
	// 健康检查
	nodeNum, err := global.NodesInfo.Flush(global.Landlady)
	if err != nil {
		return 2
	}

	for _, server := range global.NodesInfo.Get() {
		// 连通性测试
		conn, err := net.DialTimeout("tcp", string(server.Address), 2*time.Second)
		if err != nil {
			conn, err = net.DialTimeout("tcp", string(server.Address), 2*time.Second)
			if err != nil {
				global.NodesInfo.Del(server.ID)
				global.NodeNum.Dec()
				nodeNum--
			}
		}
		conn.Close()
	}
	if nodeNum <= global.NodeNum.GetPreNum()/2 {
		return 2
	} else if nodeNum < global.NodeNum.GetPreNum() {
		return 1
	}
	return 0
}
