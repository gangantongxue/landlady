package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Setup test测试前置函数
// 1. 开始时切换工作目录到cmd/目录下
// 2. 结束时切换到测试文件所在目录
func Setup(t *testing.T) {
	// 保存原始工作目录，用于测试结束后恢复
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}

	// 测试结束时恢复原始工作目录
	t.Cleanup(func() {
		if err = os.Chdir(originalDir); err != nil {
			t.Errorf("恢复工作目录失败: %v", err)
		}
	})

	// 获取当前测试函数所在文件的路径
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("无法获取测试文件路径")
	}

	// 计算主函数目录（cmd/）
	testDir := filepath.Dir(filename)
	targetDir := filepath.Join(testDir, "..")
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		t.Fatalf("获取绝对路径失败: %v", err)
	}

	// 设置工作目录
	if err := os.Chdir(absTargetDir); err != nil {
		t.Fatalf("切换到主函数目录失败: %v", err)
	}
}
