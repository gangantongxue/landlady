package opts

import (
	"os"
	"testing"

	"github.com/gangantongxue/landlady/cmd/test"
)

// TestInit_ParameterParsing 测试Init()参数解析
func TestInit_ParameterParsing(t *testing.T) {
	test.Setup(t)
	os.Args = []string{
		"test-program",
		"--node-num=5",
		"--node-id=node1",
		"--addr=127.0.0.1:8080",
		"--client-addr=127.0.0.1:8081",
		"--data-dir=/tmp/testdata",
		"--join",
		"--join-addr=127.0.0.1:8082",
	}

	o := &Opts{}
	err := Init(o)
	if err != nil {
		t.Fatalf("Init() 解析参数失败: %v", err)
	}

	if o.NodeNum != 5 {
		t.Errorf("期望NodeNum=5，实际=%d", o.NodeNum)
	}
	if o.NodeID != "node1" {
		t.Errorf("NodeID 解析错误，期望: node1，实际: %s", o.NodeID)
	}
	if o.PeerAddr != "127.0.0.1:8080" {
		t.Errorf("期望PeerAddr=127.0.0.1:8080，实际=%s", o.PeerAddr)
	}
	if o.ClientAddr != "127.0.0.1:8081" {
		t.Errorf("期望ClientAddr=127.0.0.1:8081，实际=%s", o.ClientAddr)
	}
	if o.DataDir != "/tmp/testdata" {
		t.Errorf("期望DataDir=/tmp/testdata，实际=%s", o.DataDir)
	}
	if !o.Join {
		t.Error("期望Join=true，实际=false")
	}
	if o.GuideAddr != "127.0.0.1:8082" {
		t.Errorf("期望JoinAddr=127.0.0.1:8082，实际=%s", o.GuideAddr)
	}
}
