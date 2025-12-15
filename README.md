# Landlady - 类Consul服务发现与管理工具

Landlady是一款基于Go语言开发的类Consul服务发现与管理工具，提供服务注册、发现、健康检查和动态更新等功能。

## 核心功能

### 1. 服务发现自动化
- 服务注册：支持服务自动注册和手动注册
- 服务发现：支持按名称、标签、健康状态过滤
- 健康检查：支持HTTP、TCP、gRPC和自定义脚本检查
- 动态更新：实时监控服务变化，推送更新

### 2. gRPC封装库
- 提供易用的gRPC接口，支持服务间通信
- 支持流式传输，实现实时服务监控
- 提供完整的错误处理和日志记录

### 3. 快速接入机制
- 提供简洁的库函数，帮助用户快速接入中央管理节点
- 支持服务注册、发现、健康检查等核心功能
- 提供丰富的配置选项，满足不同场景需求

## 架构设计

### 整体架构
```
┌─────────────────────────────────────────────────────────┐
│                       客户端层                           │
├─────────────────────────────────────────────────────────┤
│  服务注册客户端  │  服务发现客户端  │  管理控制台客户端     │
├─────────────────────────────────────────────────────────┤
│                       API层                              │
├─────────────────────────────────────────────────────────┤
│  gRPC API  │  HTTP API  │  内部通信API                    │
├─────────────────────────────────────────────────────────┤
│                       核心层                             │
├─────────────────────────────────────────────────────────┤
│  服务注册中心  │  服务发现引擎  │  健康检查模块  │  配置管理  │
├─────────────────────────────────────────────────────────┤
│                       存储层                             │
├─────────────────────────────────────────────────────────┤
│  Raft集群  │  状态机  │  持久化存储  │  快照管理              │
└─────────────────────────────────────────────────────────┘
```

### 核心组件

| 组件 | 主要功能 | 文件位置 |
|------|----------|----------|
| 服务注册中心 | 处理服务注册、注销、更新请求 | cmd/service/registry.go |
| 服务发现引擎 | 处理服务发现请求，支持过滤和监控 | cmd/service/discovery.go |
| 健康检查模块 | 定期检查服务健康状态 | cmd/service/health.go |
| gRPC服务器 | 提供gRPC接口，处理客户端请求 | cmd/grpc/server.go |
| gRPC客户端 | 提供简洁的API，帮助用户快速接入 | cmd/grpc/client.go |
| 中央管理节点接入库 | 提供更高级别的API封装 | cmd/landlady/client.go |
| Raft集群管理 | 实现分布式一致性 | cmd/lraft/raft.go, guide.go, join.go |
| 状态机 | 服务数据存储和同步 | cmd/model/fsm.go, service.go |

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/gangantongxue/landlady.git
cd landlady
```

### 2. 编译项目

```bash
go mod tidy
go build -o landlady ./cmd/main.go
```

### 3. 启动服务器

```bash
# 启动第一个节点（引导节点）
./landlady --node-id=node1 --peer-addr=127.0.0.1:8001 --client-addr=127.0.0.1:9001 --data-dir=./data/node1

# 启动第二个节点（加入已有集群）
./landlady --node-id=node2 --peer-addr=127.0.0.1:8002 --client-addr=127.0.0.1:9002 --data-dir=./data/node2 --join --guide-addr=127.0.0.1:8001

# 启动第三个节点（加入已有集群）
./landlady --node-id=node3 --peer-addr=127.0.0.1:8003 --client-addr=127.0.0.1:9003 --data-dir=./data/node3 --join --guide-addr=127.0.0.1:8001
```

### 4. 配置选项

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--node-id` | 节点唯一标识符 | 无 |
| `--peer-addr` | 节点间通信地址，格式：IP:PORT | 无 |
| `--client-addr` | 客户端通信地址，格式：IP:PORT | 无 |
| `--data-dir` | 数据存储目录 | `var/lib/landlady` |
| `--join` | 是否加入已有集群 | `false` |
| `--guide-addr` | 引导节点地址，格式：IP:PORT | 无 |
| `--node-num` | 集群节点数量 | `3` |
| `--health-check-interval (s)` | 健康检查间隔，单位：秒 | `5` |
| `--log-to-console` | 是否将日志输出到控制台 | `false` |
| `--log-max-size (MB)` | 日志文件最大大小，单位：MB | `100` |
| `--log-max-backups` | 保留的日志文件最大数量 | `7` |
| `--log-max-age` | 日志文件最大保留天数 | `30` |
| `--log-compress` | 是否压缩日志文件 | `true` |

### 5. 运行测试客户端

```bash
cd test
go run client.go
```

### 6. 使用客户端

#### 6.1 服务端示例

服务端在启动时向Landlady注册自己的IP、端口和名称：

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gangantongxue/landlady/cmd/landlady"
)

func main() {
    // 创建Landlady客户端
    llClient, err := landlady.NewClient(&landlady.Config{
        Addr: "127.0.0.1:9001", // Landlady中心地址
    })
    if err != nil {
        log.Fatalf("Failed to create Landlady client: %v", err)
    }
    defer llClient.Close()

    // 注册服务到Landlady
    serverPort := 8080
    serviceID, err := llClient.RegisterService(
        "hello-service",          // 服务名称
        "127.0.0.1",             // 服务地址
        serverPort,               // 服务端口
        landlady.WithTags("hello", "v1"),  // 服务标签
        landlady.WithHealthCheck("tcp", 10*time.Second, 5*time.Second), // 健康检查
    )
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }

    fmt.Printf("Service registered successfully! ServiceID: %s\n", serviceID)
    fmt.Printf("Service listening on: 127.0.0.1:%d\n", serverPort)

    // 启动HTTP服务处理请求
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from server at 127.0.0.1:%d!", serverPort)
    })

    // 服务关闭时注销服务
    defer func() {
        if err := llClient.DeregisterService(serviceID); err != nil {
            log.Printf("Failed to deregister service: %v", err)
        }
    }()

    // 启动HTTP服务器
    if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
        log.Fatalf("Failed to start HTTP server: %v", err)
    }
}
```

#### 6.2 客户端示例

客户端通过服务名称从Landlady获取服务的IP和端口，实现动态服务发现：

```go
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
    // 创建Landlady客户端
    llClient, err := landlady.NewClient(&landlady.Config{
        Addr: "127.0.0.1:9001", // Landlady中心地址
    })
    if err != nil {
        log.Fatalf("Failed to create Landlady client: %v", err)
    }
    defer llClient.Close()

    // 发现服务
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

    // 选择一个服务实例
    service := services[0]
    fmt.Printf("Found service: %s\n", service.Name)
    fmt.Printf("Service address: %s:%d\n", service.Address, service.Port)

    // 发送HTTP请求到服务
    reqURL := fmt.Sprintf("http://%s:%d/hello", service.Address, service.Port)
    resp, err := http.Get(reqURL)
    if err != nil {
        log.Fatalf("Failed to send request: %v", err)
    }
    defer resp.Body.Close()

    // 处理响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Failed to read response: %v", err)
    }

    fmt.Printf("Response from service: %s\n", string(body))

    // 监控服务变化（可选）
    updates, err := llClient.WatchService("hello-service", true)
    if err != nil {
        log.Fatalf("Failed to watch service: %v", err)
    }

    // 处理服务变化通知
    go func() {
        for serviceList := range updates {
            fmt.Printf("\nService changes detected! Current healthy services: %d\n", len(serviceList))
            for i, svc := range serviceList {
                fmt.Printf("Service %d: %s:%d\n", i+1, svc.Address, svc.Port)
            }
        }
    }()

    // 持续运行，模拟客户端持续使用服务
    fmt.Println("\nClient is running. Press Ctrl+C to exit.")
    for {
        time.Sleep(5 * time.Second)
        // 定期重新发现服务并发送请求...
    }
}
```

## API文档

### 1. 服务注册

**函数签名**：
```go
func (c *Client) RegisterService(serviceName, address string, port int, opts ...ServiceOption) (string, error)
```

**参数**：
- `serviceName`：服务名称
- `address`：服务地址
- `port`：服务端口
- `opts`：服务选项，可选

**返回值**：
- `string`：注册成功的服务ID
- `error`：错误信息，nil表示成功

**服务选项**：
```go
// WithTags 设置服务标签
func WithTags(tags ...string) ServiceOption

// WithMeta 设置服务元数据
func WithMeta(meta map[string]string) ServiceOption

// WithHealthCheck 设置健康检查
func WithHealthCheck(checkType string, interval, timeout time.Duration) ServiceOption
```

### 2. 服务注销

**函数签名**：
```go
func (c *Client) DeregisterService(serviceID string) error
```

**参数**：
- `serviceID`：要注销的服务ID

**返回值**：
- `error`：错误信息，nil表示成功

### 3. 服务发现

**函数签名**：
```go
func (c *Client) DiscoverService(serviceName string, healthyOnly bool, tags ...string) ([]*model.Service, error)
```

**参数**：
- `serviceName`：要发现的服务名称
- `healthyOnly`：是否只返回健康的服务
- `tags`：服务标签过滤，可选

**返回值**：
- `[]*model.Service`：服务列表
- `error`：错误信息，nil表示成功

### 4. 服务监控

**函数签名**：
```go
func (c *Client) WatchService(serviceName string, healthyOnly bool, tags ...string) (<-chan []*model.Service, error)
```

**参数**：
- `serviceName`：要监控的服务名称
- `healthyOnly`：是否只监控健康的服务
- `tags`：服务标签过滤，可选

**返回值**：
- `<-chan []*model.Service`：服务更新通道，当服务有变化时会发送服务列表
- `error`：错误信息，nil表示成功

### 5. 健康检查

**函数签名**：
```go
func (c *Client) CheckHealth(serviceID string) (string, error)
```

**参数**：
- `serviceID`：要检查的服务ID

**返回值**：
- `string`：健康状态，可能的值：
  - `passing`：服务健康
  - `warning`：服务警告
  - `critical`：服务不健康
- `error`：错误信息，nil表示成功

### 6. 关闭客户端

**函数签名**：
```go
func (c *Client) Close() error
```

**返回值**：
- `error`：错误信息，nil表示成功

## 开发计划

### 第一阶段：核心功能开发（已完成）
- 服务注册中心
- 服务发现引擎
- 健康检查模块
- 状态机扩展

### 第二阶段：gRPC封装库开发（已完成）
- gRPC服务协议定义
- gRPC服务器实现
- gRPC客户端实现

### 第三阶段：快速接入库开发（已完成）
- 中央管理节点接入库
- 高级API封装

### 第四阶段：测试与优化（进行中）
- 单元测试
- 集成测试
- 性能优化

### 第五阶段：文档完善（进行中）
- API文档
- 使用示例
- 部署指南

## 后续规划

1. 支持多数据中心部署
2. 支持服务网格集成
3. 支持更丰富的健康检查类型
4. 支持服务路由和负载均衡
5. 提供Web管理控制台
6. 集成监控和告警功能

## 贡献

欢迎贡献代码、报告bug或提出建议！以下是贡献指南：

### 1. 报告Bug

在GitHub Issues中报告bug时，请包含以下信息：
- 清晰的bug描述
- 复现步骤
- 预期结果
- 实际结果
- 操作系统和Go版本
- 相关日志和错误信息

### 2. 提交代码

#### 2.1 Fork仓库

首先，Fork Landlady仓库到自己的GitHub账号。

#### 2.2 克隆仓库

```bash
git clone https://github.com/your-username/landlady.git
cd landlady
```

#### 2.3 创建分支

创建一个新的分支来开发新功能或修复bug：

```bash
git checkout -b feature/your-feature-name
```

#### 2.4 开发和测试

开发新功能或修复bug，并确保所有测试通过：

```bash
go test ./...
```

#### 2.5 提交代码

提交代码时，请使用清晰的提交信息：

```bash
git add .
git commit -m "feat: add new feature"
```

提交信息格式：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码风格更新（不影响功能）
- `refactor`: 代码重构（不影响功能）
- `test`: 添加或更新测试
- `chore`: 构建或工具更新

#### 2.6 推送到远程仓库

```bash
git push origin feature/your-feature-name
```

#### 2.7 创建Pull Request

在GitHub上创建Pull Request，描述你的更改，并链接相关的Issue。

### 3. 代码风格要求

- 遵循Go标准代码风格
- 使用`go fmt`格式化代码
- 使用`golint`检查代码质量
- 编写清晰的注释
- 添加单元测试

### 4. 开发流程

1. 提出问题或功能建议
2. 讨论并确定解决方案
3. 实现功能或修复bug
4. 编写测试
5. 提交Pull Request
6. 代码审查
7. 合并代码

## 示例应用

### 1. 服务注册示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/gangantongxue/landlady/cmd/landlady"
)

func main() {
    // 创建客户端
    client, err := landlady.NewClient(&landlady.Config{
        Addr: "127.0.0.1:9001",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 注册服务
    serviceID, err := client.RegisterService(
        "user-service",
        "127.0.0.1",
        8080,
        landlady.WithTags("user", "v1"),
        landlady.WithMeta(map[string]string{
            "version": "1.0.0",
            "env":     "production",
        }),
        landlady.WithHealthCheck("http", 10*time.Second, 5*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("服务注册成功，ServiceID: %s\n", serviceID)

    // 等待中断信号
    select {}
}
```

### 2. 服务发现示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/gangantongxue/landlady/cmd/landlady"
)

func main() {
    // 创建客户端
    client, err := landlady.NewClient(&landlady.Config{
        Addr: "127.0.0.1:9001",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 发现服务
    services, err := client.DiscoverService("user-service", true)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("发现 %d 个健康服务：\n", len(services))
    for _, service := range services {
        fmt.Printf("- 服务名称: %s, 地址: %s:%d, 版本: %s\n", 
            service.Name, service.Address, service.Port, service.Meta["version"])
    }
}
```

### 3. 服务监控示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/gangantongxue/landlady/cmd/landlady"
)

func main() {
    // 创建客户端
    client, err := landlady.NewClient(&landlady.Config{
        Addr: "127.0.0.1:9001",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 监控服务变化
    updates, err := client.WatchService("user-service", true)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("开始监控服务变化...")
    for services := range updates {
        fmt.Printf("服务更新：发现 %d 个健康服务\n", len(services))
        for _, service := range services {
            fmt.Printf("- 服务名称: %s, 地址: %s:%d\n", service.Name, service.Address, service.Port)
        }
    }
}

## 许可证

MIT License
