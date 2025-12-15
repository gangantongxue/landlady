package model

import (
	"encoding/json"
	"github.com/gangantongxue/ggl"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// Tenant 租户数据
type Tenant map[string]string

// ServiceStore 服务存储
type ServiceStore struct {
	Services     map[string]*Service          // 服务映射，key: serviceID
	HealthChecks map[string]*HealthCheck      // 健康检查映射，key: checkID
	ServiceIndex map[string][]*Service        // 服务名称索引，key: serviceName
}

// FSM 状态机
type FSM struct {
	mu    sync.RWMutex
	store Tenant
	services ServiceStore // 服务存储
}

// Apply 应用日志
func (f *FSM) Apply(l *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 解析命令
	var cmd map[string]interface{}
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		ggl.Error("Apply logger failed :", ggl.Err(err))
		return nil
	}
	
	op, ok := cmd["op"].(string)
	if !ok {
		ggl.Error("Apply logger failed : op is not string")
		return nil
	}
	
	// 初始化服务存储（如果尚未初始化）
	if f.services.Services == nil {
		f.services = ServiceStore{
			Services:     make(map[string]*Service),
			HealthChecks: make(map[string]*HealthCheck),
			ServiceIndex: make(map[string][]*Service),
		}
	}
	
	switch op {
	// 原有命令
	case "set":
		k, _ := cmd["k"].(string)
		v, _ := cmd["v"].(string)
		f.store[k] = v
		
	case "del":
		k, _ := cmd["k"].(string)
		delete(f.store, k)
		
	// 服务相关命令
	case "register_service":
		f.applyRegisterService(cmd)
		
	case "deregister_service":
		f.applyDeregisterService(cmd)
		
	case "update_service":
		f.applyUpdateService(cmd)
		
	case "register_health_check":
		f.applyRegisterHealthCheck(cmd)
		
	case "deregister_health_check":
		f.applyDeregisterHealthCheck(cmd)
		
	case "update_health_status":
		f.applyUpdateHealthStatus(cmd)
		
	default:
		ggl.Error("Apply logger failed : unknown op")
	}
	return nil
}

// applyRegisterService 应用服务注册命令
func (f *FSM) applyRegisterService(cmd map[string]interface{}) {
	// 解析服务信息
	serviceID, _ := cmd["service_id"].(string)
	serviceName, _ := cmd["name"].(string)
	address, _ := cmd["address"].(string)
	port, _ := cmd["port"].(float64)
	
	tags := make([]string, 0)
	if tagsInterface, ok := cmd["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}
	
	meta := make(map[string]string)
	if metaInterface, ok := cmd["meta"].(map[string]interface{}); ok {
		for k, v := range metaInterface {
			if vStr, ok := v.(string); ok {
				meta[k] = vStr
			}
		}
	}
	
	// 创建服务
	service := &Service{
		ID:        serviceID,
		Name:      serviceName,
		Address:   address,
		Port:      int(port),
		Tags:      tags,
		Meta:      meta,
		Health:    true,
		Timestamp: time.Now(),
	}
	
	// 存储服务
	f.services.Services[serviceID] = service
	
	// 更新服务索引
	if _, ok := f.services.ServiceIndex[serviceName]; !ok {
		f.services.ServiceIndex[serviceName] = make([]*Service, 0)
	}
	f.services.ServiceIndex[serviceName] = append(f.services.ServiceIndex[serviceName], service)
}

// applyDeregisterService 应用服务注销命令
func (f *FSM) applyDeregisterService(cmd map[string]interface{}) {
	serviceID, _ := cmd["service_id"].(string)
	
	// 获取服务
	service, ok := f.services.Services[serviceID]
	if !ok {
		return
	}
	
	// 从服务索引中删除
	if services, ok := f.services.ServiceIndex[service.Name]; ok {
		var newServices []*Service
		for _, s := range services {
			if s.ID != serviceID {
				newServices = append(newServices, s)
			}
		}
		f.services.ServiceIndex[service.Name] = newServices
	}
	
	// 删除服务
	delete(f.services.Services, serviceID)
	
	// 删除相关健康检查
	for checkID, check := range f.services.HealthChecks {
		if check.ServiceID == serviceID {
			delete(f.services.HealthChecks, checkID)
		}
	}
}

// applyUpdateService 应用服务更新命令
func (f *FSM) applyUpdateService(cmd map[string]interface{}) {
	serviceID, _ := cmd["service_id"].(string)
	
	// 获取服务
	service, ok := f.services.Services[serviceID]
	if !ok {
		return
	}
	
	// 更新服务信息
	if name, ok := cmd["name"].(string); ok {
		// 如果服务名称改变，需要更新索引
		if name != service.Name {
			// 从旧索引中删除
			if services, ok := f.services.ServiceIndex[service.Name]; ok {
				var newServices []*Service
				for _, s := range services {
					if s.ID != serviceID {
						newServices = append(newServices, s)
					}
				}
				f.services.ServiceIndex[service.Name] = newServices
			}
			
			// 添加到新索引
			if _, ok := f.services.ServiceIndex[name]; !ok {
				f.services.ServiceIndex[name] = make([]*Service, 0)
			}
			f.services.ServiceIndex[name] = append(f.services.ServiceIndex[name], service)
			
			// 更新服务名称
			service.Name = name
		}
	}
	
	if address, ok := cmd["address"].(string); ok {
		service.Address = address
	}
	
	if port, ok := cmd["port"].(float64); ok {
		service.Port = int(port)
	}
	
	if tagsInterface, ok := cmd["tags"].([]interface{}); ok {
		tags := make([]string, 0)
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
		service.Tags = tags
	}
	
	if metaInterface, ok := cmd["meta"].(map[string]interface{}); ok {
		meta := make(map[string]string)
		for k, v := range metaInterface {
			if vStr, ok := v.(string); ok {
				meta[k] = vStr
			}
		}
		service.Meta = meta
	}
	
	service.Timestamp = time.Now()
}

// applyRegisterHealthCheck 应用健康检查注册命令
func (f *FSM) applyRegisterHealthCheck(cmd map[string]interface{}) {
	checkID, _ := cmd["check_id"].(string)
	serviceID, _ := cmd["service_id"].(string)
	checkType, _ := cmd["type"].(string)
	
	// 解析配置
	config := make(map[string]string)
	if configInterface, ok := cmd["config"].(map[string]interface{}); ok {
		for k, v := range configInterface {
			if vStr, ok := v.(string); ok {
				config[k] = vStr
			}
		}
	}
	
	// 解析时间参数
	interval := time.Second * 10
	if intervalStr, ok := cmd["interval"].(float64); ok {
		interval = time.Duration(intervalStr) * time.Second
	}
	
	timeout := time.Second * 5
	if timeoutStr, ok := cmd["timeout"].(float64); ok {
		timeout = time.Duration(timeoutStr) * time.Second
	}
	
	deregisterAfter := time.Minute * 5
	if deregisterAfterStr, ok := cmd["deregister_after"].(float64); ok {
		deregisterAfter = time.Duration(deregisterAfterStr) * time.Second
	}
	
	// 创建健康检查
	check := &HealthCheck{
		ID:             checkID,
		ServiceID:      serviceID,
		Type:           checkType,
		Config:         config,
		Interval:       interval,
		Timeout:        timeout,
		DeregisterCriticalServiceAfter: deregisterAfter,
		LastCheck:      time.Now(),
		Status:         "passing",
	}
	
	// 存储健康检查
	f.services.HealthChecks[checkID] = check
	
	// 更新服务健康状态
	if service, ok := f.services.Services[serviceID]; ok {
		service.Health = true
	}
}

// applyDeregisterHealthCheck 应用健康检查注销命令
func (f *FSM) applyDeregisterHealthCheck(cmd map[string]interface{}) {
	checkID, _ := cmd["check_id"].(string)
	
	// 删除健康检查
	delete(f.services.HealthChecks, checkID)
}

// applyUpdateHealthStatus 应用健康状态更新命令
func (f *FSM) applyUpdateHealthStatus(cmd map[string]interface{}) {
	serviceID, _ := cmd["service_id"].(string)
	status, _ := cmd["status"].(string)
	
	// 获取服务
	service, ok := f.services.Services[serviceID]
	if !ok {
		return
	}
	
	// 更新服务健康状态
	service.Health = status == "passing"
	
	// 更新相关健康检查状态
	for _, check := range f.services.HealthChecks {
		if check.ServiceID == serviceID {
			check.Status = status
			check.LastCheck = time.Now()
		}
	}
}

// Get 获取值
func (f *FSM) Get(k string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	v, ok := f.store[k]
	return v, ok
}

// GetService 获取服务
func (f *FSM) GetService(serviceID string) (*Service, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	service, ok := f.services.Services[serviceID]
	return service, ok
}

// GetServices 获取所有服务
func (f *FSM) GetServices() []*Service {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var services []*Service
	for _, service := range f.services.Services {
		services = append(services, service)
	}
	return services
}

// GetServicesByName 按名称获取服务
func (f *FSM) GetServicesByName(name string) []*Service {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if services, ok := f.services.ServiceIndex[name]; ok {
		return services
	}
	return []*Service{}
}

// GetHealthCheck 获取健康检查
func (f *FSM) GetHealthCheck(checkID string) (*HealthCheck, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	check, ok := f.services.HealthChecks[checkID]
	return check, ok
}

// GetHealthChecksByServiceID 按服务ID获取健康检查
func (f *FSM) GetHealthChecksByServiceID(serviceID string) []*HealthCheck {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var checks []*HealthCheck
	for _, check := range f.services.HealthChecks {
		if check.ServiceID == serviceID {
			checks = append(checks, check)
		}
	}
	return checks
}

// Snapshot 快照
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// 复制服务存储
	var servicesCopy ServiceStore
	if f.services.Services != nil {
		servicesCopy = ServiceStore{
			Services:     make(map[string]*Service),
			HealthChecks: make(map[string]*HealthCheck),
			ServiceIndex: make(map[string][]*Service),
		}
		
		// 复制服务
		for k, v := range f.services.Services {
			servicesCopy.Services[k] = v
		}
		
		// 复制健康检查
		for k, v := range f.services.HealthChecks {
			servicesCopy.HealthChecks[k] = v
		}
		
		// 复制服务索引
		for k, v := range f.services.ServiceIndex {
			servicesCopy.ServiceIndex[k] = v
		}
	}
	
	// 复制原有存储
	snapshot := make(Tenant)
	for k, v := range f.store {
		snapshot[k] = v
	}
	
	return &Snapshot{
		store:    snapshot,
		services: servicesCopy,
	}, nil
}

// Restore 恢复
func (f *FSM) Restore(r io.ReadCloser) error {
	defer r.Close()
	
	// 解析快照
	var snapshot struct {
		Store    Tenant       `json:"store"`
		Services ServiceStore `json:"services"`
	}
	
	if err := json.NewDecoder(r).Decode(&snapshot); err != nil {
		return err
	}
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 恢复存储
	f.store = snapshot.Store
	
	// 恢复服务存储
	if snapshot.Services.Services != nil {
		f.services = snapshot.Services
	} else {
		f.services = ServiceStore{
			Services:     make(map[string]*Service),
			HealthChecks: make(map[string]*HealthCheck),
			ServiceIndex: make(map[string][]*Service),
		}
	}
	
	return nil
}

// Snapshot 快照
type Snapshot struct {
	store    Tenant
	services ServiceStore
}

// Persist 持久化
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	
	// 构建快照数据
	snapshotData := struct {
		Store    Tenant       `json:"store"`
		Services ServiceStore `json:"services"`
	}{
		Store:    s.store,
		Services: s.services,
	}
	
	// 序列化快照数据
	bytes, err := json.Marshal(snapshotData)
	if err != nil {
		sink.Cancel()
		return err
	}
	
	// 写入快照
	if _, err := sink.Write(bytes); err != nil {
		sink.Cancel()
		return err
	}
	
	return nil
}

// Release 释放
func (s *Snapshot) Release() {
	s.store = nil
	s.services = ServiceStore{}
}
