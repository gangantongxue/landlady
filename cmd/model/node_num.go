package model

import "sync"

// NodeNum 节点数量
type NodeNum struct {
	// preNum 预定节点数量
	preNum int

	// curNum 当前节点数量
	curNum int

	// m 节点数量读写锁
	m sync.Mutex
}

// Inc 节点数量增加
func (n *NodeNum) Inc() {
	n.m.Lock()
	defer n.m.Unlock()
	n.curNum++
}

// Dec 节点数量减少
func (n *NodeNum) Dec() {
	n.m.Lock()
	defer n.m.Unlock()
	n.curNum--
}

// SetCurNum 设置当前节点数量
func (n *NodeNum) SetCurNum(num int) {
	n.m.Lock()
	defer n.m.Unlock()
	n.curNum = num
}

// SetPreNum 设置预定节点数量
func (n *NodeNum) SetPreNum(num int) {
	n.m.Lock()
	defer n.m.Unlock()
	n.preNum = num
}

// IsFull 是否已满
func (n *NodeNum) IsFull() bool {
	return n.curNum >= n.preNum
}

// GetCurNum 获取当前节点数量
func (n *NodeNum) GetCurNum() int {
	n.m.Lock()
	defer n.m.Unlock()
	return n.curNum
}

// GetPreNum 获取预定节点数量
func (n *NodeNum) GetPreNum() int {
	n.m.Lock()
	defer n.m.Unlock()
	return n.preNum
}
