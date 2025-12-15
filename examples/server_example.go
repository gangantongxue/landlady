package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gangantongxue/landlady/cmd/landlady"
)

func main() {
	// 1. 创建Landlady客户端
	llClient, err := landlady.NewClient(&landlady.Config{
		Addr: "127.0.0.1:9001", // Landlady中心地址
	})
	if err != nil {
		log.Fatalf("Failed to create Landlady client: %v", err)
	}
	defer llClient.Close()

	// 2. 注册服务到Landlady
	// 这里使用随机端口来模拟服务节点变动
	serverPort := 8080
	serviceID, err := llClient.RegisterService(
		"hello-service",                  // 服务名称
		"127.0.0.1",                      // 服务地址
		serverPort,                       // 服务端口
		landlady.WithTags("hello", "v1"), // 服务标签
		landlady.WithHealthCheck("tcp", 10*time.Second, 5*time.Second), // 健康检查
		landlady.WithMeta(map[string]string{
			"version": "1.0.0",
			"env":     "dev",
		}), // 服务元数据
	)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("Service registered successfully! ServiceID: %s\n", serviceID)
	fmt.Printf("Service listening on: 127.0.0.1:%d\n", serverPort)

	// 3. 启动HTTP服务处理请求
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from server at 127.0.0.1:%d!", serverPort)
	})

	// 4. 服务关闭时注销服务
	defer func() {
		if err := llClient.DeregisterService(serviceID); err != nil {
			log.Printf("Failed to deregister service: %v", err)
		} else {
			fmt.Printf("Service deregistered successfully! ServiceID: %s\n", serviceID)
		}
	}()

	// 5. 启动HTTP服务器
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", serverPort),
	}

	// 启动HTTP服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 保持进程运行
	fmt.Println("Server is running. Press Ctrl+C to exit.")
	select {}
}
