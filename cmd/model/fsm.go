package model

import (
	"encoding/json"
	"github.com/gangantongxue/ggl"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

// FSM 状态机
type FSM struct {
	mu    sync.RWMutex
	store Tenant
}

// Apply 应用日志
func (f *FSM) Apply(l *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	var cmd map[string]string
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		ggl.Error("Apply logger failed :", ggl.Err(err))
		return nil
	}
	switch cmd["op"] {
	case "set":
		f.store[cmd["k"]] = cmd["v"]
	case "del":
		delete(f.store, cmd["k"])
	default:
		ggl.Error("Apply logger failed : unknown op")
	}
	return nil
}

// Get 获取值
func (f *FSM) Get(k string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	v, ok := f.store[k]
	return v, ok
}

// Snapshot 快照
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	snapshot := make(Tenant)
	for k, v := range f.store {
		snapshot[k] = v
	}
	return &Snapshot{store: snapshot}, nil
}

// Restore 恢复
func (f *FSM) Restore(r io.ReadCloser) error {
	defer r.Close()
	var snapshot Tenant
	if err := json.NewDecoder(r).Decode(&snapshot); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store = snapshot
	return nil
}

// Snapshot 快照
type Snapshot struct {
	store Tenant
}

// Persist 持久化
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	bytes, _ := json.Marshal(s.store)
	if _, err := sink.Write(bytes); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

// Release 释放
func (s *Snapshot) Release() {
	s.store = nil
}
