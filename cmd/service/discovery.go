package service

import (
	"sync"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/model"
)

// DiscoverService 发现服务
func DiscoverService(name string, filters ...model.Filter) ([]*model.Service, error) {
	// 从状态机获取所有服务
	services := global.FSM.GetServices()
	
	// 按名称过滤
	nameFilter := &model.ByNameFilter{Name: name}
	services = nameFilter.Apply(services)
	
	// 应用其他过滤器
	for _, filter := range filters {
		services = filter.Apply(services)
	}
	
	return services, nil
}

// WatchService 监控服务变化
type Watcher struct {
	name     string
	filters  []model.Filter
	updates  chan []*model.Service
	stopCh   chan struct{}
	wg       sync.WaitGroup
	lastSeen map[string]time.Time
}

// NewWatcher 创建服务监控器
func NewWatcher(name string, filters ...model.Filter) *Watcher {
	return &Watcher{
		name:     name,
		filters:  filters,
		updates:  make(chan []*model.Service, 10),
		stopCh:   make(chan struct{}),
		lastSeen: make(map[string]time.Time),
	}
}

// Start 启动监控
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watch()
}

// Stop 停止监控
func (w *Watcher) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	close(w.updates)
}

// Updates 返回服务更新通道
func (w *Watcher) Updates() <-chan []*model.Service {
	return w.updates
}

// watch 监控服务变化
func (w *Watcher) watch() {
	defer w.wg.Done()
	
	// 初始检查
	w.checkServices()
	
	// 定期检查
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.checkServices()
		}
	}
}

// checkServices 检查服务变化
func (w *Watcher) checkServices() {
	// 发现服务
	services, err := DiscoverService(w.name, w.filters...)
	if err != nil {
		return
	}
	
	// 检查服务是否有变化
	hasChanged := false
	current := make(map[string]time.Time)
	
	for _, service := range services {
		current[service.ID] = service.Timestamp
		
		// 检查服务是否新增或更新
		if lastSeen, ok := w.lastSeen[service.ID]; !ok || service.Timestamp.After(lastSeen) {
			hasChanged = true
		}
	}
	
	// 检查服务是否删除
	for id := range w.lastSeen {
		if _, ok := current[id]; !ok {
			hasChanged = true
			break
		}
	}
	
	// 如果有变化，发送更新
	if hasChanged {
		w.lastSeen = current
		select {
		case w.updates <- services:
		default:
			// 通道已满，丢弃
		}
	}
}

// GetService 获取单个服务
func GetService(serviceID string) (*model.Service, bool) {
	return global.FSM.GetService(serviceID)
}

// GetAllServices 获取所有服务
func GetAllServices() []*model.Service {
	return global.FSM.GetServices()
}

// GetServicesByHealth 根据健康状态获取服务
func GetServicesByHealth(healthy bool) []*model.Service {
	services := global.FSM.GetServices()
	healthFilter := &model.ByHealthFilter{Healthy: healthy}
	return healthFilter.Apply(services)
}

// GetServicesByTag 根据标签获取服务
func GetServicesByTag(tag string) []*model.Service {
	services := global.FSM.GetServices()
	tagFilter := &model.ByTagFilter{Tag: tag}
	return tagFilter.Apply(services)
}