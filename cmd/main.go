package main

import (
	"fmt"
	"path/filepath"

	"github.com/gangantongxue/ggl"
	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/grpc"
	"github.com/gangantongxue/landlady/cmd/lraft"
	"github.com/gangantongxue/landlady/cmd/opts"
)

func main() {
	cfg := &ggl.Config{
		LogFileName:   "landlady_log_2006-01-02.log",
		LogFileDir:    filepath.Join(global.Opts.DataDir, "log"),
		LogMaxSize:    global.Opts.LogMaxSize,
		LogMaxBackups: global.Opts.LogMaxBackups,
		LogMaxAge:     global.Opts.LogMaxAge,
		LogCompress:   global.Opts.LogCompress,
		ToConsole:     global.Opts.ToConsole,
	}
	d := ggl.New(cfg)
	defer d.Stop()

	if err := opts.Init(global.Opts); err != nil {
		fmt.Printf("opts init failed: %v\n", err)
		return
	}
	if err := lraft.Init(); err != nil {
		fmt.Printf("raft init failed: %v\n", err)
		return
	}

	// 启动gRPC服务器
	grpcServer := grpc.NewServer(global.Opts.ClientAddr)
	go func() {
		if err := grpcServer.Start(); err != nil {
			fmt.Printf("gRPC server start failed: %v\n", err)
		}
	}()
	defer grpcServer.Stop()

	lraft.InitHealthCheck()

	// 阻塞主线程，防止程序退出
	select {}
}
