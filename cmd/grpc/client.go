package grpc

import (
	"context"
	"time"

	"github.com/gangantongxue/landlady/cmd/model"
	"github.com/gangantongxue/landlady/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn    *grpc.ClientConn
	client  proto.LandladyServiceClient
	watcher proto.WatchServiceClient
}

// NewClient 创建gRPC客户端
func NewClient(addr string) (*Client, error) {
	// 连接gRPC服务器
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// 创建客户端
	return &Client{
		conn:    conn,
		client:  proto.NewLandladyServiceClient(conn),
		watcher: proto.NewWatchServiceClient(conn),
	}, nil
}

// Close 关闭gRPC客户端
func (c *Client) Close() error {
	return c.conn.Close()
}

// RegisterService 服务注册
func (c *Client) RegisterService(service *model.Service, healthCheck *model.HealthCheck) (string, error) {
	// 转换服务信息
	protoService := &proto.Service{
		Id:        service.ID,
		Name:      service.Name,
		Address:   service.Address,
		Port:      int32(service.Port),
		Tags:      service.Tags,
		Meta:      service.Meta,
		Health:    service.Health,
		Timestamp: service.Timestamp.Unix(),
	}

	// 转换健康检查配置
	var protoHealthCheck *proto.HealthCheck
	if healthCheck != nil {
		protoHealthCheck = &proto.HealthCheck{
			Id:              healthCheck.ID,
			ServiceId:       healthCheck.ServiceID,
			Type:            healthCheck.Type,
			Config:          healthCheck.Config,
			Interval:        int64(healthCheck.Interval.Seconds()),
			Timeout:         int64(healthCheck.Timeout.Seconds()),
			DeregisterAfter: int64(healthCheck.DeregisterCriticalServiceAfter.Seconds()),
			LastCheck:       healthCheck.LastCheck.Unix(),
			Status:          healthCheck.Status,
		}
	}

	// 发送注册请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.RegisterService(ctx, &proto.RegisterServiceRequest{
		Service:     protoService,
		HealthCheck: protoHealthCheck,
	})
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", err
	}

	return resp.ServiceId, nil
}

// DeregisterService 服务注销
func (c *Client) DeregisterService(serviceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.DeregisterService(ctx, &proto.DeregisterServiceRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return err
	}

	return nil
}

// UpdateService 服务更新
func (c *Client) UpdateService(service *model.Service) error {
	// 转换服务信息
	protoService := &proto.Service{
		Id:        service.ID,
		Name:      service.Name,
		Address:   service.Address,
		Port:      int32(service.Port),
		Tags:      service.Tags,
		Meta:      service.Meta,
		Health:    service.Health,
		Timestamp: service.Timestamp.Unix(),
	}

	// 发送更新请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.UpdateService(ctx, &proto.UpdateServiceRequest{
		Service: protoService,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return err
	}

	return nil
}

// DiscoverService 服务发现
func (c *Client) DiscoverService(name string, healthyOnly bool, tags ...string) ([]*model.Service, error) {
	// 发送发现请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.DiscoverService(ctx, &proto.DiscoverServiceRequest{
		Name:        name,
		HealthyOnly: healthyOnly,
		Tags:        tags,
	})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, err
	}

	// 转换服务信息
	services := make([]*model.Service, len(resp.Services))
	for i, s := range resp.Services {
		services[i] = &model.Service{
			ID:        s.Id,
			Name:      s.Name,
			Address:   s.Address,
			Port:      int(s.Port),
			Tags:      s.Tags,
			Meta:      s.Meta,
			Health:    s.Health,
			Timestamp: time.Unix(s.Timestamp, 0),
		}
	}

	return services, nil
}

// CheckHealth 健康检查
func (c *Client) CheckHealth(serviceID string) (string, error) {
	// 发送健康检查请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.CheckHealth(ctx, &proto.CheckHealthRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", err
	}

	return resp.Status, nil
}

// WatchService 监控服务变化
func (c *Client) WatchService(name string, healthyOnly bool, tags ...string) (<-chan []*model.Service, error) {
	// 发送监控请求
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := c.watcher.WatchService(ctx, &proto.WatchServiceRequest{
		Name:        name,
		HealthyOnly: healthyOnly,
		Tags:        tags,
	})
	if err != nil {
		cancel()
		return nil, err
	}

	// 创建结果通道
	updates := make(chan []*model.Service, 10)

	// 读取监控结果
	go func() {
		defer cancel()
		defer close(updates)

		for {
			resp, err := stream.Recv()
			if err != nil {
				break
			}

			// 转换服务信息
			services := make([]*model.Service, len(resp.Services))
			for i, s := range resp.Services {
				services[i] = &model.Service{
					ID:        s.Id,
					Name:      s.Name,
					Address:   s.Address,
					Port:      int(s.Port),
					Tags:      s.Tags,
					Meta:      s.Meta,
					Health:    s.Health,
					Timestamp: time.Unix(s.Timestamp, 0),
				}
			}

			// 发送结果
			select {
			case updates <- services:
			case <-ctx.Done():
				return
			}
		}
	}()

	return updates, nil
}
