package grpc

import (
	"context"
	"net"
	"time"

	"github.com/gangantongxue/landlady/cmd/model"
	"github.com/gangantongxue/landlady/cmd/service"
	"github.com/gangantongxue/landlady/proto"
	"google.golang.org/grpc"
)

// Server gRPC服务器
type Server struct {
	proto.UnimplementedLandladyServiceServer
	proto.UnimplementedWatchServiceServer
	addr   string
	server *grpc.Server
}

// NewServer 创建gRPC服务器
func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

// Start 启动gRPC服务器
func (s *Server) Start() error {
	// 监听端口
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// 创建gRPC服务器
	s.server = grpc.NewServer()

	// 注册服务
	proto.RegisterLandladyServiceServer(s.server, s)
	proto.RegisterWatchServiceServer(s.server, s)

	// 启动服务器
	return s.server.Serve(ln)
}

// Stop 停止gRPC服务器
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// RegisterService 服务注册
func (s *Server) RegisterService(ctx context.Context, req *proto.RegisterServiceRequest) (*proto.RegisterServiceResponse, error) {
	// 转换服务信息
	svc := &model.Service{
		ID:        req.Service.Id,
		Name:      req.Service.Name,
		Address:   req.Service.Address,
		Port:      int(req.Service.Port),
		Tags:      req.Service.Tags,
		Meta:      req.Service.Meta,
		Health:    req.Service.Health,
		Timestamp: time.Unix(req.Service.Timestamp, 0),
	}

	// 注册服务
	if err := service.RegisterService(svc); err != nil {
		return &proto.RegisterServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 如果有健康检查配置，注册健康检查
	if req.HealthCheck != nil {
		healthCheck := &model.HealthCheck{
			ID:                             req.HealthCheck.Id,
			ServiceID:                      svc.ID,
			Type:                           req.HealthCheck.Type,
			Config:                         req.HealthCheck.Config,
			Interval:                       time.Duration(req.HealthCheck.Interval) * time.Second,
			Timeout:                        time.Duration(req.HealthCheck.Timeout) * time.Second,
			DeregisterCriticalServiceAfter: time.Duration(req.HealthCheck.DeregisterAfter) * time.Second,
			LastCheck:                      time.Unix(req.HealthCheck.LastCheck, 0),
			Status:                         req.HealthCheck.Status,
		}
		if err := service.RegisterHealthCheck(healthCheck); err != nil {
			return &proto.RegisterServiceResponse{
				Success:   false,
				Message:   err.Error(),
				ServiceId: svc.ID,
			}, nil
		}
	}

	return &proto.RegisterServiceResponse{
		Success:   true,
		Message:   "service registered successfully",
		ServiceId: svc.ID,
	}, nil
}

// DeregisterService 服务注销
func (s *Server) DeregisterService(ctx context.Context, req *proto.DeregisterServiceRequest) (*proto.DeregisterServiceResponse, error) {
	// 注销服务
	if err := service.DeregisterService(req.ServiceId); err != nil {
		return &proto.DeregisterServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &proto.DeregisterServiceResponse{
		Success: true,
		Message: "service deregistered successfully",
	}, nil
}

// UpdateService 服务更新
func (s *Server) UpdateService(ctx context.Context, req *proto.UpdateServiceRequest) (*proto.UpdateServiceResponse, error) {
	// 转换服务信息
	svc := &model.Service{
		ID:        req.Service.Id,
		Name:      req.Service.Name,
		Address:   req.Service.Address,
		Port:      int(req.Service.Port),
		Tags:      req.Service.Tags,
		Meta:      req.Service.Meta,
		Health:    req.Service.Health,
		Timestamp: time.Unix(req.Service.Timestamp, 0),
	}

	// 更新服务
	if err := service.UpdateService(svc); err != nil {
		return &proto.UpdateServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &proto.UpdateServiceResponse{
		Success: true,
		Message: "service updated successfully",
	}, nil
}

// DiscoverService 服务发现
func (s *Server) DiscoverService(ctx context.Context, req *proto.DiscoverServiceRequest) (*proto.DiscoverServiceResponse, error) {
	// 构建过滤器
	filters := make([]model.Filter, 0)

	// 按健康状态过滤
	if req.HealthyOnly {
		filters = append(filters, &model.ByHealthFilter{Healthy: true})
	}

	// 按标签过滤
	for _, tag := range req.Tags {
		filters = append(filters, &model.ByTagFilter{Tag: tag})
	}

	// 发现服务
	services, err := service.DiscoverService(req.Name, filters...)
	if err != nil {
		return &proto.DiscoverServiceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换服务信息
	protoServices := make([]*proto.Service, len(services))
	for i, svc := range services {
		protoServices[i] = &proto.Service{
			Id:        svc.ID,
			Name:      svc.Name,
			Address:   svc.Address,
			Port:      int32(svc.Port),
			Tags:      svc.Tags,
			Meta:      svc.Meta,
			Health:    svc.Health,
			Timestamp: svc.Timestamp.Unix(),
		}
	}

	return &proto.DiscoverServiceResponse{
		Success:  true,
		Message:  "service discovery successful",
		Services: protoServices,
	}, nil
}

// CheckHealth 健康检查
func (s *Server) CheckHealth(ctx context.Context, req *proto.CheckHealthRequest) (*proto.CheckHealthResponse, error) {
	// 获取服务
	service, ok := service.GetService(req.ServiceId)
	if !ok {
		return &proto.CheckHealthResponse{
			Success: false,
			Message: "service not found",
		}, nil
	}

	// 获取健康状态
	status := "passing"
	if !service.Health {
		status = "critical"
	}

	return &proto.CheckHealthResponse{
		Success: true,
		Message: "health check successful",
		Status:  status,
	}, nil
}

// WatchService 监控服务变化
func (s *Server) WatchService(req *proto.WatchServiceRequest, stream proto.WatchService_WatchServiceServer) error {
	// 构建过滤器
	filters := make([]model.Filter, 0)

	// 按健康状态过滤
	if req.HealthyOnly {
		filters = append(filters, &model.ByHealthFilter{Healthy: true})
	}

	// 按标签过滤
	for _, tag := range req.Tags {
		filters = append(filters, &model.ByTagFilter{Tag: tag})
	}

	// 创建服务监控器
	watcher := service.NewWatcher(req.Name, filters...)
	watcher.Start()
	defer watcher.Stop()

	// 发送服务更新
	for services := range watcher.Updates() {
		// 转换服务信息
		protoServices := make([]*proto.Service, len(services))
		for i, s := range services {
			protoServices[i] = &proto.Service{
				Id:        s.ID,
				Name:      s.Name,
				Address:   s.Address,
				Port:      int32(s.Port),
				Tags:      s.Tags,
				Meta:      s.Meta,
				Health:    s.Health,
				Timestamp: s.Timestamp.Unix(),
			}
		}

		// 发送更新
		if err := stream.Send(&proto.WatchServiceResponse{Services: protoServices}); err != nil {
			return err
		}
	}

	return nil
}
