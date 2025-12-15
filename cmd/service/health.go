package service

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/model"
	"google.golang.org/grpc"
)

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check(service *model.Service) (string, error)
}

// HTTPHealthChecker HTTP健康检查器
type HTTPHealthChecker struct {
	client *http.Client
}

// NewHTTPHealthChecker 创建HTTP健康检查器
func NewHTTPHealthChecker() *HTTPHealthChecker {
	return &HTTPHealthChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Check 执行HTTP健康检查
func (c *HTTPHealthChecker) Check(service *model.Service) (string, error) {
	url := "http://" + service.Address + ":" + string(rune(service.Port)) + "/health"
	resp, err := c.client.Get(url)
	if err != nil {
		return "critical", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "passing", nil
	}
	return "critical", nil
}

// TCPHealthChecker TCP健康检查器
type TCPHealthChecker struct{}

// NewTCPHealthChecker 创建TCP健康检查器
func NewTCPHealthChecker() *TCPHealthChecker {
	return &TCPHealthChecker{}
}

// Check 执行TCP健康检查
func (c *TCPHealthChecker) Check(service *model.Service) (string, error) {
	addr := service.Address + ":" + string(rune(service.Port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "critical", err
	}
	defer conn.Close()
	return "passing", nil
}

// GRPCHealthChecker gRPC健康检查器
type GRPCHealthChecker struct {
	client *grpc.ClientConn
}

// NewGRPCHealthChecker 创建gRPC健康检查器
func NewGRPCHealthChecker() *GRPCHealthChecker {
	return &GRPCHealthChecker{}
}

// Check 执行gRPC健康检查
func (c *GRPCHealthChecker) Check(service *model.Service) (string, error) {
	addr := service.Address + ":" + string(rune(service.Port))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "critical", err
	}
	defer conn.Close()

	// 简单检查连接是否成功
	return "passing", nil
}

// ScriptHealthChecker 脚本健康检查器
type ScriptHealthChecker struct{}

// NewScriptHealthChecker 创建脚本健康检查器
func NewScriptHealthChecker() *ScriptHealthChecker {
	return &ScriptHealthChecker{}
}

// Check 执行脚本健康检查
func (c *ScriptHealthChecker) Check(service *model.Service) (string, error) {
	// 从服务元数据中获取脚本路径
	script, ok := service.Meta["health_check_script"]
	if !ok {
		return "warning", nil
	}

	cmd := exec.Command(script, service.Address, string(rune(service.Port)))
	err := cmd.Run()
	if err != nil {
		return "critical", err
	}

	return "passing", nil
}

// HealthCheckManager 健康检查管理器
type HealthCheckManager struct {
	checkers map[string]HealthChecker
	running  bool
}

// NewHealthCheckManager 创建健康检查管理器
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checkers: map[string]HealthChecker{
			"http":   NewHTTPHealthChecker(),
			"tcp":    NewTCPHealthChecker(),
			"grpc":   NewGRPCHealthChecker(),
			"script": NewScriptHealthChecker(),
		},
		running: false,
	}
}

// Start 启动健康检查
func (m *HealthCheckManager) Start() {
	if m.running {
		return
	}
	m.running = true

	// 启动健康检查协程
	go m.run()
}

// Stop 停止健康检查
func (m *HealthCheckManager) Stop() {
	m.running = false
}

// run 执行健康检查
func (m *HealthCheckManager) run() {
	// 定期执行健康检查
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for m.running {
		select {
		case <-ticker.C:
			m.checkAllServices()
		}
	}
}

// checkAllServices 检查所有服务健康状态
func (m *HealthCheckManager) checkAllServices() {
	// 获取所有服务
	services := global.FSM.GetServices()

	// 检查每个服务
	for _, service := range services {
		go m.checkService(service)
	}
}

// checkService 检查单个服务健康状态
func (m *HealthCheckManager) checkService(service *model.Service) {
	// 从服务元数据获取健康检查类型
	checkType, ok := service.Meta["health_check_type"]
	if !ok {
		checkType = "tcp" // 默认TCP检查
	}

	// 获取健康检查器
	checker, ok := m.checkers[checkType]
	if !ok {
		checker = m.checkers["tcp"] // 默认TCP检查
	}

	// 执行健康检查
	status, err := checker.Check(service)
	if err != nil {
		status = "critical"
	}

	// 更新服务健康状态
	_ = UpdateHealthStatus(service.ID, status)
}
