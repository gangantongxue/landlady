package main

import (
	"fmt"
	"io"
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

	// 2. 发现服务
	// 只需要知道服务名称，不需要硬编码IP和端口
	services, err := llClient.DiscoverService(
		"hello-service", // 服务名称
		true,            // 只获取健康的服务
		"hello", "v1",  // 服务标签过滤
	)
	if err != nil {
		log.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) == 0 {
		log.Fatalf("No healthy services found")
	}

	// 3. 选择一个服务实例（这里简单选择第一个）
	service := services[0]
	fmt.Printf("Found service: %s\n", service.Name)
	fmt.Printf("Service address: %s:%d\n", service.Address, service.Port)

	// 4. 发送HTTP请求到服务
	reqURL := fmt.Sprintf("http://%s:%d/hello", service.Address, service.Port)
	resp, err := http.Get(reqURL)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 5. 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Response from service: %s\n", string(body))

	// 6. 监控服务变化（可选）
	fmt.Println("\nStart watching service changes...")
	updates, err := llClient.WatchService(
		"hello-service", // 服务名称
		true,            // 只监控健康的服务
	)
	if err != nil {
		log.Fatalf("Failed to watch service: %v", err)
	}

	// 7. 处理服务变化通知
	go func() {
		for serviceList := range updates {
			fmt.Printf("\nService changes detected! Current healthy services: %d\n", len(serviceList))
			for i, svc := range serviceList {
				fmt.Printf("Service %d: %s:%d\n", i+1, svc.Address, svc.Port)
			}
		}
	}()

	// 8. 持续运行，模拟客户端持续使用服务
	fmt.Println("\nClient is running. Press Ctrl+C to exit.")
	for {
		// 定期重新发现服务并发送请求
		time.Sleep(5 * time.Second)

		// 重新发现服务
		services, err := llClient.DiscoverService("hello-service", true)
		if err != nil {
			log.Printf("Failed to rediscover service: %v", err)
			continue
		}

		if len(services) == 0 {
			log.Println("No healthy services found")
			continue
		}

		// 选择一个服务实例
		service := services[0]
		reqURL := fmt.Sprintf("http://%s:%d/hello", service.Address, service.Port)
		
		// 发送请求
		resp, err := http.Get(reqURL)
		if err != nil {
			log.Printf("Failed to send request: %v", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("Updated response: %s\n", string(body))
	}
}