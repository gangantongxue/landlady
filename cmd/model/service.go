package model

import (
	"time"
)

// Service 服务信息
type Service struct {
	ID        string            // 服务唯一标识
	Name      string            // 服务名称
	Address   string            // 服务地址
	Port      int               // 服务端口
	Tags      []string          // 服务标签
	Meta      map[string]string // 服务元数据
	Health    bool              // 健康状态
	Timestamp time.Time         // 注册时间
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	ID             string            // 健康检查ID
	ServiceID      string            // 关联的服务ID
	Type           string            // 检查类型：http, tcp, grpc, script
	Config         map[string]string // 检查配置
	Interval       time.Duration     // 检查间隔
	Timeout        time.Duration     // 检查超时
	DeregisterCriticalServiceAfter time.Duration // 服务不健康多久后注销
	LastCheck      time.Time         // 上次检查时间
	Status         string            // 当前状态：passing, warning, critical
}

// Filter 服务发现过滤器
type Filter interface {
	Apply(services []*Service) []*Service
}

// ByNameFilter 按名称过滤
type ByNameFilter struct {
	Name string
}

func (f *ByNameFilter) Apply(services []*Service) []*Service {
	var result []*Service
	for _, s := range services {
		if s.Name == f.Name {
			result = append(result, s)
		}
	}
	return result
}

// ByHealthFilter 按健康状态过滤
type ByHealthFilter struct {
	Healthy bool
}

func (f *ByHealthFilter) Apply(services []*Service) []*Service {
	var result []*Service
	for _, s := range services {
		if s.Health == f.Healthy {
			result = append(result, s)
		}
	}
	return result
}

// ByTagFilter 按标签过滤
type ByTagFilter struct {
	Tag string
}

func (f *ByTagFilter) Apply(services []*Service) []*Service {
	var result []*Service
	for _, s := range services {
		for _, tag := range s.Tags {
			if tag == f.Tag {
				result = append(result, s)
				break
			}
		}
	}
	return result
}