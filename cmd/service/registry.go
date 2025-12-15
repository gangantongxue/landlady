package service

import (
	"encoding/json"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
	"github.com/gangantongxue/landlady/cmd/model"
)

// RegisterService 注册服务
func RegisterService(service *model.Service) error {
	// 生成服务ID（如果未提供）
	if service.ID == "" {
		service.ID = service.Name + "_" + time.Now().Format("20060102150405") + "_" + service.Address + ":" + string(rune(service.Port))
	}

	// 设置默认值
	if service.Meta == nil {
		service.Meta = make(map[string]string)
	}
	service.Timestamp = time.Now()
	service.Health = true

	// 构建注册命令
	cmd := map[string]interface{}{
		"op":         "register_service",
		"service_id": service.ID,
		"name":       service.Name,
		"address":    service.Address,
		"port":       service.Port,
		"tags":       service.Tags,
		"meta":       service.Meta,
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}

// DeregisterService 注销服务
func DeregisterService(serviceID string) error {
	// 构建注销命令
	cmd := map[string]interface{}{
		"op":         "deregister_service",
		"service_id": serviceID,
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}

// UpdateService 更新服务
func UpdateService(service *model.Service) error {
	// 构建更新命令
	cmd := map[string]interface{}{
		"op":         "update_service",
		"service_id": service.ID,
		"name":       service.Name,
		"address":    service.Address,
		"port":       service.Port,
		"tags":       service.Tags,
		"meta":       service.Meta,
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}

// RegisterHealthCheck 注册健康检查
func RegisterHealthCheck(check *model.HealthCheck) error {
	// 生成检查ID（如果未提供）
	if check.ID == "" {
		check.ID = check.ServiceID + "_health_check_" + time.Now().Format("20060102150405")
	}

	// 设置默认值
	if check.Config == nil {
		check.Config = make(map[string]string)
	}
	if check.Interval == 0 {
		check.Interval = 10 * time.Second
	}
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}
	if check.DeregisterCriticalServiceAfter == 0 {
		check.DeregisterCriticalServiceAfter = 5 * time.Minute
	}
	check.LastCheck = time.Now()
	check.Status = "passing"

	// 构建注册命令
	cmd := map[string]interface{}{
		"op":               "register_health_check",
		"check_id":         check.ID,
		"service_id":       check.ServiceID,
		"type":             check.Type,
		"config":           check.Config,
		"interval":         check.Interval.Seconds(),
		"timeout":          check.Timeout.Seconds(),
		"deregister_after": check.DeregisterCriticalServiceAfter.Seconds(),
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}

// DeregisterHealthCheck 注销健康检查
func DeregisterHealthCheck(checkID string) error {
	// 构建注销命令
	cmd := map[string]interface{}{
		"op":       "deregister_health_check",
		"check_id": checkID,
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}

// UpdateHealthStatus 更新健康状态
func UpdateHealthStatus(serviceID, status string) error {
	// 构建更新命令
	cmd := map[string]interface{}{
		"op":         "update_health_status",
		"service_id": serviceID,
		"status":     status,
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// 应用Raft日志
	if err := global.Landlady.Apply(cmdBytes, time.Second).Error(); err != nil {
		return err
	}

	return nil
}
