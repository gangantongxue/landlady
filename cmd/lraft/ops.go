package lraft

import (
	"encoding/json"
	"time"

	"github.com/gangantongxue/landlady/cmd/global"
)

// Set 设置值
func Set(k, v string) error {
	cmd := map[string]string{
		"op": "set",
		"k":  k,
		"v":  v,
	}
	data, _ := json.Marshal(cmd)
	future := global.Landlady.Apply(data, time.Second)
	return future.Error()
}

// Get 获取值
func Get(k string) (string, bool) {
	return global.FSM.Get(k)
}
