package version

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type versionInfo struct {
	Version    string `yaml:"version"`
	Author     string `yaml:"author"`
	Email      string `yaml:"email"`
	Repository string `yaml:"repository"`
}

var version versionInfo

// Init 初始化版本信息
// 从version.yaml中读取版本信息
func Init() {
	yamlPath := filepath.Join("version", "version.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Panic("read version file error", err)
	}
	err = yaml.Unmarshal(data, &version)
	if err != nil {
		log.Panic("unmarshal version file error", err)
	}
}

// GetVersion 获取版本信息
func GetVersion() string {
	return version.Version
}

// GetAuthor 获取作者信息
func GetAuthor() string {
	return version.Author
}

// GetEmail 获取邮箱信息
func GetEmail() string {
	return version.Email
}

// GetRepository 获取仓库信息
func GetRepository() string {
	return version.Repository
}

// PrintVersion 打印版本信息
func PrintVersion() {
	versionStr := fmt.Sprintf("landlady version: %s,\nauthor: %s,\nrepository: %s,\nemail: %s,\nif you have any question, please contact me.",
		version.Version, version.Author, version.Repository, version.Email)
	fmt.Println(versionStr)
}
