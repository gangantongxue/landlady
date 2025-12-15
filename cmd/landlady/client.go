package landlady

import (
	"time"

	"github.com/gangantongxue/landlady/cmd/grpc"
	"github.com/gangantongxue/landlady/cmd/model"
)

// Config 客户端配置
type Config struct {
	Addr string // 中央管理节点地址
}

// Client 中央管理节点客户端
type Client struct {
	grpcClient *grpc.Client
}

// NewClient 创建中央管理节点客户端
func NewClient(config *Config) (*Client, error) {
	// 创建gRPC客户端
	grpcClient, err := grpc.NewClient(config.Addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		grpcClient: grpcClient,
	}, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.grpcClient.Close()
}

// ServiceOption 服务选项
type ServiceOption func(*model.Service)

// WithTags 设置服务标签
func WithTags(tags ...string) ServiceOption {
	return func(s *model.Service) {
		s.Tags = tags
	}
}

// WithMeta 设置服务元数据
func WithMeta(meta map[string]string) ServiceOption {
	return func(s *model.Service) {
		s.Meta = meta
	}
}

// WithHealthCheck 设置健康检查
func WithHealthCheck(checkType string, interval, timeout time.Duration) ServiceOption {
	return func(s *model.Service) {
		if s.Meta == nil {
			s.Meta = make(map[string]string)
		}
		s.Meta["health_check_type"] = checkType
	}
}

// RegisterService 注册服务
func (c *Client) RegisterService(serviceName, address string, port int, opts ...ServiceOption) (string, error) {
	// 创建服务信息
	service := &model.Service{
		Name:      serviceName,
		Address:   address,
		Port:      port,
		Health:    true,
		Timestamp: time.Now(),
		Tags:      []string{},
		Meta:      make(map[string]string),
	}

	// 应用服务选项
	for _, opt := range opts {
		opt(service)
	}

	// 注册服务
	return c.grpcClient.RegisterService(service, nil)
}

// DeregisterService 注销服务
func (c *Client) DeregisterService(serviceID string) error {
	return c.grpcClient.DeregisterService(serviceID)
}

// DiscoverService 发现服务
func (c *Client) DiscoverService(serviceName string, healthyOnly bool, tags ...string) ([]*model.Service, error) {
	return c.grpcClient.DiscoverService(serviceName, healthyOnly, tags...)
}

// WatchService 监控服务变化
func (c *Client) WatchService(serviceName string, healthyOnly bool, tags ...string) (<-chan []*model.Service, error) {
	return c.grpcClient.WatchService(serviceName, healthyOnly, tags...)
}

// CheckHealth 检查服务健康状态
func (c *Client) CheckHealth(serviceID string) (string, error) {
	return c.grpcClient.CheckHealth(serviceID)
}
