非常好的问题，这个在生产环境中很常见。
在 HashiCorp Raft 中，“动态加入节点” 是支持的，但要分清两种情况：

---

## ✅ 一、前提：集群已启动并有一个 Leader

Raft 的所有成员变更（Add/Remove）**必须通过当前 Leader 节点** 来完成。

也就是说，新节点 **不能自己加入集群**，必须由 **Leader 发起 `AddVoter()` 或 `AddNonvoter()` 操作**。

---

## ✅ 二、标准流程（动态添加新节点）

下面以 Go 语言（`hashicorp/raft` 包）为例。

假设现在有：

* 一个正在运行的 Raft 集群（3 个节点）
* 新节点的地址是：`127.0.0.1:9003`

### **1️⃣ 新节点启动 Raft，但不要马上参与投票**

新节点需要先以 **non-voter**（观察者）身份加入网络，以防止因为日志不一致而破坏 quorum。

```go
config := raft.DefaultConfig()
config.LocalID = raft.ServerID("node3")

addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:9003")
transport, _ := raft.NewTCPTransport("127.0.0.1:9003", addr, 3, 10*time.Second, os.Stderr)

// Storage
snapshots, _ := raft.NewFileSnapshotStore("/tmp/raft-node3", 2, os.Stderr)
logStore := raft.NewInmemStore()
stableStore := raft.NewInmemStore()

r, err := raft.NewRaft(config, myFSM, logStore, stableStore, snapshots, transport)
if err != nil {
    panic(err)
}
```

此时节点启动了 Raft 实例，但还**没被集群认可**。

---

### **2️⃣ 由 Leader 节点发起添加请求**

Leader 调用：

```go
future := raftInstance.AddVoter(
    raft.ServerID("node3"),
    raft.ServerAddress("127.0.0.1:9003"),
    0, // prevIndex: 一般填 0
    10*time.Second,
)
if err := future.Error(); err != nil {
    log.Printf("failed to add voter: %v", err)
}
```

👉 这一步会在集群内复制新的成员配置日志条目，等到日志提交后，集群配置更新完成。

---

### **3️⃣ 新节点自动同步日志**

添加后，Leader 会自动将完整日志同步给 node3，等它追上日志进度后，它就能正式参与选举。

---

## ✅ 三、节点类型说明

| 类型       | 方法              | 可投票 | 用途                |
| -------- | --------------- | --- | ----------------- |
| Voter    | `AddVoter()`    | ✅ 是 | 正式节点              |
| Nonvoter | `AddNonvoter()` | ❌ 否 | 日志追赶用或即将升级为 voter |

👉 一般我们先调用 `AddNonvoter()`，同步完日志后再调用 `AddVoter()` 进行升级。

---

## ✅ 四、如果原有节点已经宕机怎么办？

如果宕机节点无法恢复（比如磁盘坏了），则需要：

1. 先通过 Leader 调用 `RemoveServer()` 移除宕机节点：

   ```go
   future := raftInstance.RemoveServer("node2", 0, 0)
   if err := future.Error(); err != nil {
       log.Printf("failed to remove node2: %v", err)
   }
   ```
2. 再执行上面的 `AddVoter()` 添加新节点。

这样可以保持集群成员数恒定（例如 3 节点集群）。

---

## ✅ 五、Leader 宕机怎么办？

如果 Leader 宕机了：

* 剩余节点会自动重新选出新的 Leader。
* 你要向**新的 Leader** 发起 `AddVoter()` 或 `RemoveServer()`。
* 新节点本身无法自行加入。

---

## ✅ 六、常见坑点

| 问题                       | 说明                                    |
| ------------------------ | ------------------------------------- |
| ❌ 新节点直接参与投票              | 容易导致 quorum 不可达                       |
| ❌ 节点地址写错                 | Raft 的 ServerID/Address 必须一致，否则日志同步失败 |
| ❌ 从 follower 发起 AddVoter | 会失败，必须是 leader 执行                     |
| ❌ 新节点没清空旧快照              | 会报 “logs inconsistent” 错误             |

---

## ✅ 七、总结

动态加入节点的正确流程：

```
新节点启动 -> Leader 调用 AddNonvoter() -> 日志追上 -> Leader 调用 AddVoter()
```

若替换宕机节点：

```
Leader 调用 RemoveServer(旧节点)
Leader 调用 AddVoter(新节点)
```

---

是否希望我给出一个 **完整可运行的 Go 示例**（包括 Leader + Follower + 新节点动态加入）？
我可以写成三个独立进程文件（main_leader.go、main_follower.go、main_join.go）。
